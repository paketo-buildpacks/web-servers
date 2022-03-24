package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testYarnFrontend(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when building a Yarn frontend app using NGINX as the webserver", func() {
		var (
			image     occam.Image
			container occam.Container

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "yarn-nginx-javascript-frontend"))
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("creates a working OCI image", func() {
			var err error
			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithBuildpacks(webServersBuildpack).
				WithPullPolicy("never").
				WithEnv(map[string]string{"BP_NODE_RUN_SCRIPTS": "build"}).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String())

			Expect(logs).To(ContainLines(ContainSubstring("Node Engine Buildpack")))
			Expect(logs).To(ContainLines(ContainSubstring("Yarn Buildpack")))
			Expect(logs).To(ContainLines(ContainSubstring("Yarn Install Buildpack")))
			Expect(logs).To(ContainLines(ContainSubstring("Node Run Script Buildpack")))
			Expect(logs).To(ContainLines(ContainSubstring("Nginx Server Buildpack")))
			Expect(logs).NotTo(ContainLines(ContainSubstring("HTTP Server Buildpack")))
			Expect(logs).NotTo(ContainLines(ContainSubstring("Procfile Buildpack")))

			container, err = docker.Container.Run.
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())
			Eventually(container).Should(Serve(ContainSubstring("<title>React App</title>")).OnPort(8080).WithEndpoint("/index.html"))
		})
	})

	context("when building a Yarn frontend app using HTTPD as the webserver", func() {
		var (
			image     occam.Image
			container occam.Container

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "yarn-httpd-javascript-frontend"))
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("creates a working OCI image", func() {
			var err error
			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithBuildpacks(webServersBuildpack).
				WithPullPolicy("never").
				WithEnv(map[string]string{"BP_NODE_RUN_SCRIPTS": "build"}).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String())

			Expect(logs).To(ContainLines(ContainSubstring("Node Engine Buildpack")))
			Expect(logs).To(ContainLines(ContainSubstring("Yarn Buildpack")))
			Expect(logs).To(ContainLines(ContainSubstring("Yarn Install Buildpack")))
			Expect(logs).To(ContainLines(ContainSubstring("Node Run Script Buildpack")))
			Expect(logs).NotTo(ContainLines(ContainSubstring("Nginx Server Buildpack")))
			Expect(logs).To(ContainLines(ContainSubstring("HTTP Server Buildpack")))
			Expect(logs).NotTo(ContainLines(ContainSubstring("Procfile Buildpack")))

			container, err = docker.Container.Run.
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())
			Eventually(container).Should(Serve(ContainSubstring("<title>React App</title>")).OnPort(8080).WithEndpoint("/index.html"))
		})
	})
}
