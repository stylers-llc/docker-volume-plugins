package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	mountedvolume "mounted-volume"

	"github.com/dchest/uniuri"
	"github.com/docker/go-plugins-helpers/volume"
)

type gfsDriver struct {
	servers   []string
	gfsVolume string
	mountedvolume.Driver
}

func (p *gfsDriver) Validate(req *volume.CreateRequest) error {

	_, serversDefinedInOpts := req.Options["servers"]
	_, glusteroptsInOpts := req.Options["glusteropts"]

	if len(p.servers) > 0 && (serversDefinedInOpts || glusteroptsInOpts) {
		return fmt.Errorf("SERVERS is set, options are not allowed")
	}
	if serversDefinedInOpts && glusteroptsInOpts {
		return fmt.Errorf("servers is set, glusteropts are not allowed")
	}
	if len(p.servers) == 0 && !serversDefinedInOpts && !glusteroptsInOpts {
		return fmt.Errorf("One of SERVERS, driver_opts.servers or driver_opts.glusteropts must be specified")
	}

	return nil
}

func (p *gfsDriver) MountOptions(req *volume.CreateRequest) []string {

	servers, serversDefinedInOpts := req.Options["servers"]
	glusteropts, _ := req.Options["glusteropts"]

	var args []string

	if len(p.servers) > 0 {
		for _, server := range p.servers {
			args = append(args, "-s", server)
		}
		args = p.AppendVolumeOptionsByVolumeName(args, req.Name)
	} else if serversDefinedInOpts {
		for _, server := range strings.Split(servers, ",") {
			args = append(args, "-s", server)
		}
		args = p.AppendVolumeOptionsByVolumeName(args, req.Name)
	} else {
		args = strings.Split(glusteropts, " ")
	}

	return args
}

func (p *gfsDriver) GetSubdirArg(args []string) (int, string) {
	key := -1
	subDir := ""

	for k, v := range args {
		if strings.HasPrefix(v, "--subdir-mount") {
			key = k
			subDir = strings.Replace(v, "--subdir-mount=", "", 1)
			break
		}
	}

	return key, subDir
}

func (p *gfsDriver) PreMount(req *volume.MountRequest, args []string) error {
	tmpArgs := make([]string, len(args))
	copy(tmpArgs, args)

	removable, subDir := p.GetSubdirArg(args)

	log.Println(req.Name)
	log.Println(req)
	log.Println(removable)
	log.Println(subDir)

	if removable >= 0 {
		tmpArgs = append(tmpArgs[:removable], tmpArgs[removable+1:]...)
	} else {
		return nil
	}

	tmpDir := "tmp/gfs-tmp-" + uniuri.New()

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("error creating tmp directory %s: %s", tmpDir, err.Error())
	}

	p.UnMountTmpDir(tmpDir)

	tmpArgs[len(tmpArgs)-1] = tmpDir

	cmd := exec.Command(p.Driver.MountExecutable, tmpArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Command output: %s\n", out)
		fmt.Printf("error mounting %s: %s", tmpDir, err.Error())
	}

	if err := os.MkdirAll(tmpDir+subDir, 0755); err != nil {
		return fmt.Errorf("error creating tmp directory %s: %s", tmpDir+subDir, err.Error())
	}

	p.UnMountTmpDir(tmpDir)

	return nil
}

func (p *gfsDriver) UnMountTmpDir(tmpDir string) error {
	umountCmd := exec.Command("umount", tmpDir)
	umountOut, umountErr := umountCmd.CombinedOutput()
	if umountErr != nil {
		fmt.Printf("umountCmd output:\n%s\n", string(umountOut))
		return fmt.Errorf("umountCmd failed with %s\n", umountErr.Error())
	}

	return nil
}

func (p *gfsDriver) PostMount(req *volume.MountRequest) {
}

// AppendVolumeOptionsByVolumeName appends the command line arguments into the current argument list given the volume name
func (p *gfsDriver) AppendVolumeOptionsByVolumeName(args []string, volumeName string) []string {
	args = append(args, "--volfile-id="+p.gfsVolume)
	args = append(args, "--subdir-mount=/"+volumeName)

	return args
}

func GetVolumeOptionsByVolumeName(volumeName string) []string {
	parts := strings.SplitN(volumeName, "/", 2)

	return parts
}

func buildDriver() *gfsDriver {
	var servers []string
	if os.Getenv("SERVERS") != "" {
		servers = strings.Split(os.Getenv("SERVERS"), ",")
	}
	var gfsVolume string
	gfsVolume = os.Getenv("GFS-VOLUME")

	d := &gfsDriver{
		Driver:    *mountedvolume.NewDriver("glusterfs", true, "gfs", "local"),
		servers:   servers,
		gfsVolume: gfsVolume,
	}
	d.Init(d)
	return d
}

func main() {
	log.SetFlags(0)
	d := buildDriver()
	defer d.Close()
	d.ServeUnix()
}
