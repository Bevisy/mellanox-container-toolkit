package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"runtime/debug"

	"github.com/opencontainers/runtime-spec/specs-go"
)

const (
	INFIDIBAND_DEV_DIR = "/dev/infiniband"
)

var (
	debugflag = flag.Bool("debug", false, "enable debug output")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  prestart\n        run the prestart hook\n")
	fmt.Fprintf(os.Stderr, "  poststart\n        no-op\n")
	fmt.Fprintf(os.Stderr, "  poststop\n        no-op\n")
}

func exit() {
	if err := recover(); err != nil {
		if _, ok := err.(runtime.Error); ok {
			log.Println(err)
		}
		if *debugflag {
			log.Printf("%s", debug.Stack())
		}
		os.Exit(1)
	}
	os.Exit(0)
}

func doPrestart() {
	defer exit()
	log.SetFlags(0)

	// 获取容器配置 config.json
	containerState := getContainerConfig()
	fread, err := os.OpenFile(path.Join(containerState.Bundle, "config.json"), os.O_RDONLY, 0644)
	if err != nil {
		log.Panicln("could not open OCI spec:", err)
	}
	defer fread.Close()

	var spec specs.Spec
	// 解析容器 spec
	if err = json.NewDecoder(fread).Decode(&spec); err != nil {
		log.Panicln("could not decode OCI spec:", err)
	}

	fwrite, err := os.OpenFile(path.Join(containerState.Bundle, "config.json"), os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Panicln("could not open OCI spec:", err)
	}
	defer fwrite.Close()

	// 获取 infiniband devices
	devs, err := getDevices(INFIDIBAND_DEV_DIR, "")
	if err != nil {
		log.Panicln("could not open OCI spec:", err)
	}
	spec.Linux.Devices = append(spec.Linux.Devices, devs...)

	// 添加设备到 cgroup
	for _, dev := range devs {
		spec.Linux.Resources.Devices = append(spec.Linux.Resources.Devices, specs.LinuxDeviceCgroup{
			Allow:  true,
			Type:   dev.Type,
			Major:  &dev.Major,
			Minor:  &dev.Minor,
			Access: "rwm",
		})
	}

	// 写入新的容器配置 config.json
	if err = json.NewEncoder(fwrite).Encode(spec); err != nil {
		log.Panicln("could not encode OCI spec:", err)
	}
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	switch args[0] {
	case "prestart":
		doPrestart()
		os.Exit(0)
	case "poststart":
		fallthrough
	case "poststop":
		os.Exit(0)
	default:
		flag.Usage()
		os.Exit(2)
	}
}
