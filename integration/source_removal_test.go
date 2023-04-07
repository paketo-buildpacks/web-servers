package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSourceRemoval(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("Source Removal Buildpack", func() {
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

			source, err = occam.Source(filepath.Join("testdata", "npm-nginx-javascript-frontend"))
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		context("with source removal environment variable set", func() {
			it("creates a working OCI image with source removed", func() {
				var err error
				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(webServersBuildpack).
					WithPullPolicy("never").
					WithEnv(map[string]string{
						"BP_NODE_RUN_SCRIPTS": "build",
						"BP_EXCLUDE_FILES":    "src",
					}).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())

				container, err = docker.Container.Run.
					WithEntrypoint("ls").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				logs, err = docker.Container.Logs.Execute(container.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(logs.String()).NotTo(ContainSubstring("src"))
			})
		})

		context("with no source removal environment variables set", func() {
			it("creates a working OCI image with source removed", func() {
				var err error
				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(webServersBuildpack).
					WithPullPolicy("never").
					WithEnv(map[string]string{
						"BP_NODE_RUN_SCRIPTS": "build",
					}).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())

				container, err = docker.Container.Run.
					WithEntrypoint("ls").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				logs, err = docker.Container.Logs.Execute(container.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(logs.String()).To(ContainSubstring("src"))
			})
		})
	})
}
