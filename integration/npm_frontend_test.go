package integration_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testNPMFrontend(t *testing.T, context spec.G, it spec.S) {
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

	context("when building a NPM frontend app using NGINX as the webserver", func() {
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
			Expect(logs).To(ContainLines(ContainSubstring("NPM Install Buildpack")))
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

		context("when using optional utility buildpacks", func() {
			it("creates a working OCI image and uses utility buildpacks", func() {
				var err error
				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(webServersBuildpack).
					WithPullPolicy("never").
					WithEnv(map[string]string{
						"BP_NODE_RUN_SCRIPTS":    "build",
						"BPE_SOME_VARIABLE":      "some-value",
						"BP_IMAGE_LABELS":        "some-label=some-value",
						"BP_LIVE_RELOAD_ENABLED": "true",
					}).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())

				Expect(logs).To(ContainLines(ContainSubstring("Node Engine Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("NPM Install Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("Node Run Script Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("Nginx Server Buildpack")))
				Expect(logs).NotTo(ContainLines(ContainSubstring("HTTP Server Buildpack")))
				Expect(logs).NotTo(ContainLines(ContainSubstring("Procfile Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("Environment Variables Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("Image Labels Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("Watchexec Buildpack")))

				Expect(image.Buildpacks[6].Key).To(Equal("paketo-buildpacks/environment-variables"))
				Expect(image.Buildpacks[6].Layers["environment-variables"].Metadata["variables"]).To(Equal(map[string]interface{}{"SOME_VARIABLE": "some-value"}))
				Expect(image.Labels["some-label"]).To(Equal("some-value"))

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())
				Eventually(container).Should(Serve(ContainSubstring("<title>React App</title>")).OnPort(8080).WithEndpoint("/index.html"))
			})
		})

		context("when using CA certificates", func() {
			var client *http.Client

			it.Before(func() {
				var err error
				name, err = occam.RandomName()
				Expect(err).NotTo(HaveOccurred())
				source, err = occam.Source(filepath.Join("testdata", "ca_cert_apps"))
				Expect(err).NotTo(HaveOccurred())

				caCert, err := ioutil.ReadFile(filepath.Join(source, "client_certs", "ca.pem"))
				Expect(err).ToNot(HaveOccurred())

				caCertPool := x509.NewCertPool()
				caCertPool.AppendCertsFromPEM(caCert)

				cert, err := tls.LoadX509KeyPair(filepath.Join(source, "client_certs", "cert.pem"), filepath.Join(source, "client_certs", "key.pem"))
				Expect(err).ToNot(HaveOccurred())

				client = &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							RootCAs:      caCertPool,
							Certificates: []tls.Certificate{cert},
							MinVersion:   tls.VersionTLS12,
						},
					},
				}
			})

			it("builds a working OCI image with given CA cert added to trust store", func() {
				var err error
				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(webServersBuildpack).
					WithPullPolicy("never").
					WithEnv(map[string]string{"BP_NODE_RUN_SCRIPTS": "build"}).
					Execute(name, filepath.Join(source, "npm-nginx-javascript-frontend"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Node Engine Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("NPM Install Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("Node Run Script Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("CA Certificates Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("Nginx Server Buildpack")))

				container, err = docker.Container.Run.
					WithPublish("8080").
					WithEnv(map[string]string{
						"PORT":                 "8080",
						"SERVICE_BINDING_ROOT": "/bindings",
					}).
					WithVolumes(fmt.Sprintf("%s:/bindings/ca-certificates", filepath.Join(source, "binding"))).
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					cLogs, err := docker.Container.Logs.Execute(container.ID)
					Expect(err).NotTo(HaveOccurred())
					return cLogs.String()
				}).Should(
					ContainSubstring("Added 1 additional CA certificate(s) to system truststore"),
				)

				request, err := http.NewRequest("GET", fmt.Sprintf("https://localhost:%s", container.HostPort("8080")), nil)
				Expect(err).NotTo(HaveOccurred())

				var response *http.Response
				Eventually(func() error {
					var err error
					response, err = client.Do(request)
					return err
				}).Should(BeNil())
				defer response.Body.Close()

				Expect(response.StatusCode).To(Equal(http.StatusOK))

				content, err := ioutil.ReadAll(response.Body)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(content)).To(ContainSubstring("<title>React App</title>"))
			})

		})
	})

	context("when building a NPM frontend app using HTTPD as the webserver", func() {
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

			source, err = occam.Source(filepath.Join("testdata", "npm-httpd-javascript-frontend"))
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
			Expect(logs).To(ContainLines(ContainSubstring("NPM Install Buildpack")))
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

		context("when using optional utility buildpacks", func() {
			it("creates a working OCI image and uses utility buildpacks", func() {
				var err error
				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(webServersBuildpack).
					WithPullPolicy("never").
					WithEnv(map[string]string{
						"BP_NODE_RUN_SCRIPTS":    "build",
						"BPE_SOME_VARIABLE":      "some-value",
						"BP_IMAGE_LABELS":        "some-label=some-value",
						"BP_LIVE_RELOAD_ENABLED": "true",
					}).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())

				Expect(logs).To(ContainLines(ContainSubstring("Node Engine Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("NPM Install Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("Node Run Script Buildpack")))
				Expect(logs).NotTo(ContainLines(ContainSubstring("Nginx Server Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("HTTP Server Buildpack")))
				Expect(logs).NotTo(ContainLines(ContainSubstring("Procfile Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("Environment Variables Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("Image Labels Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("Watchexec Buildpack")))

				Expect(image.Buildpacks[6].Key).To(Equal("paketo-buildpacks/environment-variables"))
				Expect(image.Buildpacks[6].Layers["environment-variables"].Metadata["variables"]).To(Equal(map[string]interface{}{"SOME_VARIABLE": "some-value"}))
				Expect(image.Labels["some-label"]).To(Equal("some-value"))

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())
				Eventually(container).Should(Serve(ContainSubstring("<title>React App</title>")).OnPort(8080).WithEndpoint("/index.html"))
			})
		})

		context("when using CA certificates", func() {
			var client *http.Client

			it.Before(func() {
				var err error
				name, err = occam.RandomName()
				Expect(err).NotTo(HaveOccurred())
				source, err = occam.Source(filepath.Join("testdata", "ca_cert_apps"))
				Expect(err).NotTo(HaveOccurred())

				caCert, err := ioutil.ReadFile(filepath.Join(source, "client_certs", "ca.pem"))
				Expect(err).ToNot(HaveOccurred())

				caCertPool := x509.NewCertPool()
				caCertPool.AppendCertsFromPEM(caCert)

				cert, err := tls.LoadX509KeyPair(filepath.Join(source, "client_certs", "cert.pem"), filepath.Join(source, "client_certs", "key.pem"))
				Expect(err).ToNot(HaveOccurred())

				client = &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							RootCAs:      caCertPool,
							Certificates: []tls.Certificate{cert},
							MinVersion:   tls.VersionTLS12,
						},
					},
				}
			})

			it("builds a working OCI image with given CA cert added to trust store", func() {
				var err error
				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(webServersBuildpack).
					WithPullPolicy("never").
					WithEnv(map[string]string{"BP_NODE_RUN_SCRIPTS": "build"}).
					Execute(name, filepath.Join(source, "npm-httpd-javascript-frontend"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Node Engine Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("NPM Install Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("Node Run Script Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("CA Certificates Buildpack")))
				Expect(logs).To(ContainLines(ContainSubstring("HTTP Server Buildpack")))

				container, err = docker.Container.Run.
					WithPublish("8080").
					WithEnv(map[string]string{
						"PORT":                 "8080",
						"SERVICE_BINDING_ROOT": "/bindings",
					}).
					WithVolumes(fmt.Sprintf("%s:/bindings/ca-certificates", filepath.Join(source, "binding"))).
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					cLogs, err := docker.Container.Logs.Execute(container.ID)
					Expect(err).NotTo(HaveOccurred())
					return cLogs.String()
				}).Should(
					ContainSubstring("Added 1 additional CA certificate(s) to system truststore"),
				)

				request, err := http.NewRequest("GET", fmt.Sprintf("https://localhost:%s", container.HostPort("8080")), nil)
				Expect(err).NotTo(HaveOccurred())

				var response *http.Response
				Eventually(func() error {
					var err error
					response, err = client.Do(request)
					return err
				}).Should(BeNil())
				defer response.Body.Close()

				Expect(response.StatusCode).To(Equal(http.StatusOK))

				content, err := ioutil.ReadAll(response.Body)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(content)).To(ContainSubstring("<title>React App</title>"))
			})
		})
	})
}
