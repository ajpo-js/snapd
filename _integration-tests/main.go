// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2015 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"

	"launchpad.net/snappy/_integration-tests/helpers/autopkgtest"
	"launchpad.net/snappy/_integration-tests/helpers/build"
	"launchpad.net/snappy/_integration-tests/helpers/config"
	"launchpad.net/snappy/_integration-tests/helpers/image"
	"launchpad.net/snappy/_integration-tests/helpers/utils"
)

const (
	baseDir        = "/tmp/snappy-test"
	defaultRelease = "rolling"
	defaultChannel = "edge"
	defaultSSHPort = 22
	dataOutputDir  = "_integration-tests/data/output/"
)

var configFileName = filepath.Join(dataOutputDir, "testconfig.json")

func setupAndRunLocalTests(rootPath, baseDir, testFilter string, img image.Image) {
	// Run the tests on the latest rolling edge image.
	if imageName, err := img.UdfCreate(); err == nil {
		autopkgtest.AdtRun(rootPath, baseDir, testFilter, autopkgtest.KvmSSHOptions(imageName))
	}
}

func setupAndRunRemoteTests(rootPath, baseDir, testFilter, testbedIP string, testbedPort int) {
	utils.ExecCommand("ssh-copy-id", "-p", strconv.Itoa(testbedPort),
		"ubuntu@"+testbedIP)
	autopkgtest.AdtRun(
		rootPath, baseDir, testFilter, autopkgtest.RemoteTestbedSSHOptions(testbedIP, testbedPort))
}

func main() {
	var (
		useSnappyFromBranch = flag.Bool("snappy-from-branch", false,
			"If this flag is used, snappy will be compiled from this branch, copied to the testbed and used for the tests. Otherwise, the snappy installed with the image will be used.")
		arch = flag.String("arch", "",
			"Architecture of the test bed. Defaults to use the same architecture as the host.")
		testbedIP = flag.String("ip", "",
			"IP of the testbed. If no IP is passed, a virtual machine will be created for the test.")
		testbedPort = flag.Int("port", defaultSSHPort,
			"SSH port of the testbed. Defaults to use port "+strconv.Itoa(defaultSSHPort))
		testFilter = flag.String("filter", "",
			"Suites or tests to run, for instance MyTestSuite, MyTestSuite.FirstCustomTest or MyTestSuite.*CustomTest")
		imgRelease = flag.String("release", defaultRelease,
			"Release of the image to be built, defaults to "+defaultRelease)
		imgChannel = flag.String("channel", defaultChannel,
			"Channel of the image to be built, defaults to "+defaultChannel)
		imgRevision = flag.String("revision", "",
			"Revision of the image to be built (can be relative to the latest available revision in the given release and channel as in -1), defaults to the empty string")
		update = flag.Bool("update", false,
			"If this flag is used, the image will be updated before running the tests.")
		targetRelease = flag.String("target-release", "",
			"If the update flag is used, the image will be updated to this release before running the tests.")
		targetChannel = flag.String("target-channel", "",
			"If the update flag is used, the image will be updated to this channel before running the tests.")
		rollback = flag.Bool("rollback", false,
			"If this flag is used, the image will be updated and then rolled back before running the tests.")
	)

	flag.Parse()

	build.Assets(*useSnappyFromBranch, *arch)

	// TODO: generate the files out of the source tree. --elopio - 2015-07-15
	utils.PrepareTargetDir(dataOutputDir)
	defer os.RemoveAll(dataOutputDir)

	// TODO: pass the config as arguments to the test binaries.
	// --elopio - 2015-07-15
	cfg := config.NewConfig(
		configFileName, *imgRelease, *imgChannel, *targetRelease, *targetChannel,
		*update, *rollback)
	cfg.Write()

	rootPath := utils.RootPath()

	if *testbedIP == "" {
		img := image.NewImage(*imgRelease, *imgChannel, *imgRevision, baseDir)
		setupAndRunLocalTests(rootPath, baseDir, *testFilter, *img)
	} else {
		setupAndRunRemoteTests(rootPath, baseDir, *testFilter, *testbedIP, *testbedPort)
	}
}
