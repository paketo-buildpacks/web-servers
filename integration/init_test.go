package integration_test

import (
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

var webServersBuildpack string

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	output, err := exec.Command("bash", "-c", "../scripts/package.sh --version 1.2.3").CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), string(output))

	webServersBuildpack, err = filepath.Abs("../build/buildpackage.cnb")
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(10 * time.Second)

	format.MaxLength = 0

	suite := spec.New("Integration", spec.Parallel(), spec.Report(report.Terminal{}))
	suite("HTTPD", testHttpd)
	suite("NGINX", testNginx)
	suite("NPM Frontend", testNPMFrontend)
	suite("Yarn Frontend", testYarnFrontend)
	suite("Source Removal", testSourceRemoval)
	suite.Run(t)
}
