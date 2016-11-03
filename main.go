package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kylelemons/godebug/pretty"
	"github.com/urfave/cli"
	"log"
	"os"
	"os/exec"
)

const (
	SCRIPTS_DIR = "./scripts"
)

type Env struct {
	Project         string `json:"project"`
	GCloudProjectID string `json:"id"`
	Prefix          string `json:"prefix"`
}

type Container struct {
	Name                 string `json:"name"`
	AppName              string `json:"appName"`
	AppDir               string `json:"appDir"`
	BuildContainerScript string `json:"buildContainerScript"`
	BuildScriptArgs      string `json:"buildScriptArgs"`
	ControllerFileName   string `json:"controllerFileName"`
	DeploymentName       string `json:"deploymentName"`
	IsRegistryImage      bool   `json:"isRegistryImage"`
	Specs                struct {
		ContainerName  string `json:"containerName"`
		DiskVolumeName string `json:"diskVolumeName"`
	} `json:"specs"`
}

type Cluster struct {
	Project     string      `json:"project"` // gcloud project
	Name        string      `json:"name"`
	ClusterName string      `json:"clusterName"`
	Context     string      `json:"context"`
	URL         string      `json:"url"`
	AppEnv      string      `json:"appEnv"`
	Registry    string      `json:"registry"`
	Containers  []Container `json:"containers"`
}

func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true

	env := loadEnv()
	log.Println("-------------------------------------------------------------------------")
	log.Println("")
	log.Println("                     Project:    ", env.Project)
	log.Println("                     Project ID: ", env.GCloudProjectID)
	log.Println("                     Prefix:     ", env.Prefix)
	log.Println("")
	log.Println("-------------------------------------------------------------------------")

	var clusterName string
	var podName string
	var remoteMySQLName string
	var remoteMySQLPassword string
	var localMySQLName string
	var localFilename string

	flags := []cli.Flag{
		cli.StringFlag{
			Name:        "cluster, c",
			Value:       "",
			Usage:       "google compute engine cluster to target",
			Destination: &clusterName,
		},
		cli.StringFlag{
			Name:        "pod, p",
			Value:       "",
			Usage:       "kubernetes pod to target (use 'all' for all)",
			Destination: &podName,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "podname",
			Flags: flags,
			Usage: "return full pod name from query",
			Action: func(c *cli.Context) error {
				getClusterConfig(env, clusterName)
				getFullPodName(podName)
				return nil
			},
		},
		{
			Name:    "ship",
			Aliases: []string{"s"},
			Flags:   flags,
			Usage:   "build, push, and deploy go and relay apps",
			Action: func(c *cli.Context) error {
				cluster := getClusterConfig(env, clusterName)
				pod := podName
				if err := buildContainer(env, pod, cluster); err != nil {
					fatLog("BuildContainer", err)
					return err
				}
				if err := pushContainer(env, pod, cluster); err != nil {
					fatLog("PushContainer", err)
					return err
				}
				deployContainer(pod, cluster)
				return nil
			},
		},
		{
			Name:    "push",
			Aliases: []string{"s"},
			Flags:   flags,
			Usage:   "push containers to the gcr.io repo",
			Action: func(c *cli.Context) error {
				cluster := getClusterConfig(env, clusterName)
				pods := []string{podName}
				if podName == "all" || podName == "" {
					pods = []string{"go", "relay"}
				}
				for _, pod := range pods {
					pushContainer(env, pod, cluster)
				}
				return nil
			},
		},
		{
			Name:    "build",
			Aliases: []string{"b"},
			Flags:   flags,
			Usage:   "build docker container and push it to gcr",
			Action: func(c *cli.Context) error {
				cluster := getClusterConfig(env, clusterName)
				buildContainer(env, podName, cluster)
				// pushContainer(env, podName, cluster)
				return nil
			},
		},
		{
			Name:    "deploy",
			Aliases: []string{"d"},
			Flags:   flags,
			Usage:   "update docker container for specified kubernetes pods",
			Action: func(c *cli.Context) error {
				cluster := getClusterConfig(env, clusterName)
				deployContainer(podName, cluster)
				return nil
			},
		},
		{
			Name:    "context",
			Aliases: []string{"c", "ctx"},
			Flags:   flags,
			Usage:   "set kubectl context",
			Action: func(c *cli.Context) error {
				cluster := getClusterConfig(env, clusterName)
				setContext(cluster)
				return nil
			},
		},
		{
			Name:    "clean",
			Aliases: []string{"clean"},
			Flags:   flags,
			Usage:   "remove local docker images for the given resource name",
			Action: func(c *cli.Context) error {
				cluster := getClusterConfig(env, clusterName)
				removeImages(env, podName, cluster)
				return nil
			},
		},
		// {
		// 	Name:    "launch-cluster",
		// 	Aliases: []string{"lc"},
		// 	Flags:   flags,
		// 	Usage:   "create gce cluster",
		// 	Action: func(c *cli.Context) error {
		// 		cluster := getClusterConfig(env, clusterName)
		// 		launchCluster(cluster)
		// 		return nil
		// 	},
		// },
		// {
		// 	Name:  "destroy-cluster",
		// 	Flags: flags,
		// 	Usage: "destroy gce cluster",
		// 	Action: func(c *cli.Context) error {
		// 		cluster := getClusterConfig(env, clusterName)
		// 		destroyCluster(cluster)
		// 		return nil
		// 	},
		// },
		// {
		// 	Name:    "specs",
		// 	Aliases: []string{"sp"},
		// 	Flags:   flags,
		// 	Usage:   "generate spec files for cluster",
		// 	Action: func(c *cli.Context) error {
		// 		getClusterConfig(env, clusterName)
		// 		return nil
		// 	},
		// },
		{
			Name: "file",
			// Aliases: []string{""},
			Flags: flags,
			Usage: "download a file from a pod",
			Action: func(c *cli.Context) error {
				getClusterConfig(env, clusterName)
				// cluster := getClusterConfig(clusterName)
				// getFileFromPod(cluster, podName)
				return nil
			},
		},
		{
			Name:    "sqldump",
			Aliases: []string{"dump"},
			Flags: append(flags,
				cli.StringFlag{Name: "remotedb, rdb",
					Value:       "",
					Usage:       "remote mysql db name",
					Destination: &remoteMySQLName,
				},
				cli.StringFlag{
					Name:        "localdb, ldb",
					Value:       "",
					Usage:       "local mysql db name",
					Destination: &localMySQLName,
				},
				cli.StringFlag{
					Name:        "password, pass, P",
					Value:       "",
					Usage:       "remote mysql password",
					Destination: &remoteMySQLPassword,
				},
			),
			Usage: "dump database from cluster mysql to local",
			Action: func(c *cli.Context) error {
				cluster := getClusterConfig(env, clusterName)
				podName = "mysql"
				execMySQLDumpOnPod(remoteMySQLName, remoteMySQLPassword, podName, cluster)
				remoteFilename := fmt.Sprintf("%s.sql", remoteMySQLName)
				localFilename := fmt.Sprintf("%s_from_%s.sql", remoteMySQLName, cluster.Name)
				getFileFromPod(remoteFilename, localFilename, podName, cluster)
				reloadLocalSQL(localMySQLName, localFilename)
				return nil
			},
		},
		{
			Name:    "sqlreload",
			Aliases: []string{"reload"},
			Flags: append(flags,
				cli.StringFlag{Name: "remotedb, rdb",
					Value:       "",
					Usage:       "remote mysql db to reload",
					Destination: &remoteMySQLName,
				},
				cli.StringFlag{
					Name:        "dumpfile, df",
					Value:       "",
					Usage:       "local mysql dump filename",
					Destination: &localFilename,
				},
				cli.StringFlag{
					Name:        "password, pass, P",
					Value:       "",
					Usage:       "remote mysql password",
					Destination: &remoteMySQLPassword,
				},
			),
			Usage: "reload remote database from local mysql dump file",
			Action: func(c *cli.Context) error {
				cluster := getClusterConfig(env, clusterName)
				podName = "mysql"
				reloadRemoteSQL(cluster, podName, remoteMySQLName, remoteMySQLPassword, localFilename)
				return nil
			},
		},
	}

	app.Run(os.Args)
}

func fatLog(actionName string, err error) {
	log.Println("-------------------------------------------------------------------------")
	log.Println("")
	log.Println(fmt.Sprintf("> %s failed: %s", actionName, err))
	log.Println("")
	log.Println("-------------------------------------------------------------------------")
}

// load .tugboat.json project file into *Env
func loadEnv() *Env {
	projectFile, err := os.Open("tugboat.json")
	if err != nil {
		log.Fatal("Could not find tugboat.json - make one from tugboat.json.example")
	}
	defer projectFile.Close()

	decoder := json.NewDecoder(projectFile)
	env := &Env{}
	decoder.Decode(env)
	return env
}

func getFullPodName(podName string) string {
	out, err := exec.Command("./scripts/get_full_podname.sh", podName).Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Full pod name: %s", string(out))
	return string(out)
}

func execMySQLDumpOnPod(db string, password string, podName string, cluster *Cluster) {
	log.Println("--------------------------------------")
	log.Println("|       executing mysql dump          |")
	log.Println("--------------------------------------")
	fullPodName := getFullPodName(podName)
	cmd := exec.Command(
		"./scripts/remote_mysqldump.sh",
		fullPodName,
		password,
		db,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func getFileFromPod(remoteFilename string, localFilename, podName string, cluster *Cluster) {
	log.Println("--------------------------------------")
	log.Printf("|           downloading file          |")
	log.Println("--------------------------------------")
	fullPodName := getFullPodName(podName)
	cmd := exec.Command(
		"./scripts/download_file_from_pod.sh",
		fullPodName,
		remoteFilename,
		localFilename,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func reloadLocalSQL(db string, filename string) {
	log.Println("--------------------------------------")
	log.Printf("|           reloading mysql           |")
	log.Println("--------------------------------------")
	cmd := exec.Command("./scripts/reload_mysql.sh", db, filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func reloadRemoteSQL(cluster *Cluster, podName string, remoteDBName string, password string, dumpFilename string) {
	if cluster.AppEnv == "production" {
		log.Println("You must manually reload a production environment cluster's database")
		return
	}
	fullPodName := getFullPodName(podName)
	log.Println("--------------------------------------")
	log.Printf("|       reloading remote mysql        |")
	log.Println("--------------------------------------")
	log.Println(" > Cluster: ", cluster.Name)
	log.Println(" > File:    ", dumpFilename)
	log.Println("--------------------------------------")
	cmd := exec.Command(
		"./scripts/reload_remote_mysql.sh",
		fullPodName,
		password,
		dumpFilename,
		remoteDBName,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func buildContainer(env *Env, podName string, cluster *Cluster) error {
	log.Println("--------------------------------------")
	log.Println("|          building container        |")
	log.Println("--------------------------------------")
	for _, container := range cluster.Containers {
		if container.Name == podName {
			log.Printf("%s/%s.sh", SCRIPTS_DIR, container.BuildContainerScript)
			cmd := exec.Command(
				fmt.Sprintf("%s/%s.sh", SCRIPTS_DIR, container.BuildContainerScript),
				cluster.Context,
				container.Specs.ContainerName,
				cluster.URL,
				cluster.AppEnv,
				env.Prefix,
				env.GCloudProjectID,
				cluster.Registry,
				container.AppDir,
				container.BuildScriptArgs,
			)
			var stderr bytes.Buffer
			cmd.Stdout = os.Stdout
			// cmd.Stderr = os.Stderr
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				return errors.New(stderr.String())
			}
			return nil
		}
	}

	return errors.New(fmt.Sprintf("No app configuration found for '%s'", podName))
}

func pushContainer(env *Env, podName string, cluster *Cluster) error {
	// push container to gcr
	//log.Println("-------------------------------------------------------------------------")
	//log.Println(" push container ", podName)
	//log.Println("-------------------------------------------------------------------------")
	log.Println("--------------------------------------")
	log.Println("|          pushing container         |")
	log.Println("--------------------------------------")

	containerName, err := containerName(podName, cluster)
	if err != nil {
		return err
	}
	log.Println("container name: ", containerName)
	log.Println("project id: ", env.GCloudProjectID)

	cmd := exec.Command(
		fmt.Sprintf("%s/%s", SCRIPTS_DIR, "push_container.sh"),
		containerName,
		env.GCloudProjectID,
		cluster.Registry,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func deployContainer(podName string, cluster *Cluster) error {
	log.Println("--------------------------------------")
	log.Println("|        deploying container         |")
	log.Println("--------------------------------------")
	container, err := getContainerByName(podName, cluster)
	if err != nil {
		return err
	}

	// cmd := exec.Command("./scripts/deploy_container.sh",
	cmd := exec.Command(
		fmt.Sprintf("%s/%s", SCRIPTS_DIR, "deploy_replica_set.sh"),
		cluster.Context,
		cluster.Name,
		container.DeploymentName,
		container.AppName,
		container.Specs.ContainerName,
		cluster.Registry,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func removeImages(env *Env, name string, cluster *Cluster) error {
	log.Println("-------------------------------------------------------------------------")
	log.Println(">  remove images matching ", name)
	log.Println("-------------------------------------------------------------------------")

	containerName, err := containerName(name, cluster)
	if err != nil {
		return err
	}
	log.Println("container name: ", containerName)
	log.Println("project id: ", env.GCloudProjectID)

	cmd := exec.Command(
		fmt.Sprintf("%s/%s", SCRIPTS_DIR, "remove_docker_images.sh"),
		containerName,
		env.Prefix,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func setProject(cluster *Cluster) {
	cmd := exec.Command("./scripts/set_project.sh", cluster.Project)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func setContext(cluster *Cluster) {
	cmd := exec.Command("./scripts/set_context.sh", cluster.Context)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

//func launchCluster(cluster *Cluster) {
//	//cmd := exec.Command(
//	//	"./scripts/launch_cluster.sh",
//	//	cluster.ClusterName,
//	//	cluster.Specs.Go.Service,
//	//	cluster.Specs.Go.Controller,
//	//	cluster.Specs.Relay.Service,
//	//	cluster.Specs.Relay.Controller,
//	//	cluster.Specs.MySQL.Service,
//	//	cluster.Specs.MySQL.Controller,
//	//)
//	//cmd.Stdout = os.Stdout
//	//cmd.Stderr = os.Stderr
//	//cmd.Run()
//}
//
//func destroyCluster(cluster *Cluster) {
//	cmd := exec.Command(
//		"./scripts/destroy_cluster.sh",
//		cluster.ClusterName,
//	)
//	cmd.Stdout = os.Stdout
//	cmd.Stderr = os.Stderr
//	cmd.Run()
//}

func getClusterConfig(env *Env, clusterName string) *Cluster {
	var cluster Cluster

	// load shared spec file config
	if len(clusterName) == 0 {
		clusterName = "<missing>"
	}
	configFilePath := fmt.Sprintf("clusters/%s/config.json", clusterName)
	configFile, err := os.Open(configFilePath)
	if err != nil {
		log.Fatal("Could not find ", configFilePath)
	}
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	decoder.Decode(&cluster)

	log.Println("-------------------------------------------------------------------------")
	pretty.Print(cluster)
	log.Println("-------------------------------------------------------------------------")

	setProject(&cluster)
	setContext(&cluster)

	return &cluster
}

func getContainerByName(name string, cluster *Cluster) (Container, error) {
	var c Container
	for _, container := range cluster.Containers {
		if container.Name == name {
			return container, nil
		}
	}
	return c, errors.New("Container not found")
}

func containerName(podName string, cluster *Cluster) (string, error) {
	for _, container := range cluster.Containers {
		log.Println("container.Name = ", container.Name)
		if container.Name == podName {
			return container.Specs.ContainerName, nil
		}
	}
	return "", errors.New("Error")
}

func controllerName(podName string, cluster *Cluster) (string, error) {
	for _, container := range cluster.Containers {
		if container.Name == podName {
			return container.ControllerFileName, nil
		}
	}

	return "", errors.New("Error")
}
