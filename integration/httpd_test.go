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

func testHttpd(t *testing.T, context spec.G, it spec.S) {
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

	context("when building a basic HTTPD app", func() {
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

			source, err = occam.Source(filepath.Join("testdata", "httpd"))
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
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String())

			Expect(logs).To(ContainLines(ContainSubstring("Buildpack for Apache HTTP Server")))

			Expect(logs).NotTo(ContainLines(ContainSubstring("Buildpack for Nginx Server")))

			container, err = docker.Container.Run.
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(BeAvailable())
			Eventually(container).Should(Serve(ContainSubstring("<body>Hello World!</body>")).OnPort(8080))
		})

		context("when using optional utility buildpacks", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(source, "Procfile"), []byte("web: httpd -f /workspace/httpd.conf -k start -DFOREGROUND"), os.ModePerm)).To(Succeed())
			})
			it.After(func() {
				Expect(os.Remove(filepath.Join(source, "Procfile"))).To(Succeed())
			})

			it("creates a working OCI image and uses the Procfile start command and other utility buildpacks", func() {
				var err error
				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(webServersBuildpack).
					WithPullPolicy("never").
					WithEnv(map[string]string{
						"BPE_SOME_VARIABLE":      "some-value",
						"BP_IMAGE_LABELS":        "some-label=some-value",
						"BP_LIVE_RELOAD_ENABLED": "true",
					}).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())

				Expect(logs).To(ContainLines(ContainSubstring("Buildpack for Apache HTTP Server")))
				Expect(logs).To(ContainLines(ContainSubstring("Buildpack for Procfile")))
				Expect(logs).To(ContainLines(ContainSubstring("Buildpack for Environment Variables")))
				Expect(logs).To(ContainLines(ContainSubstring("Buildpack for Image Labels")))
				Expect(logs).To(ContainLines(ContainSubstring("Buildpack for Watchexec")))

				Expect(logs).To(ContainLines(ContainSubstring("web: httpd -f /workspace/httpd.conf -k start -DFOREGROUND")))

				Expect(image.Buildpacks[4].Key).To(Equal("paketo-buildpacks/environment-variables"))
				Expect(image.Buildpacks[4].Layers["environment-variables"].Metadata["variables"]).To(Equal(map[string]interface{}{"SOME_VARIABLE": "some-value"}))
				Expect(image.Labels["some-label"]).To(Equal("some-value"))

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())
				Eventually(container).Should(Serve(ContainSubstring("<body>Hello World!</body>")).OnPort(8080))
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
					Execute(name, filepath.Join(source, "httpd"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Buildpack for CA Certificates")))
				Expect(logs).To(ContainLines(ContainSubstring("Buildpack for Apache HTTP Server")))

				container, err = docker.Container.Run.
					WithPublish("8080").
					WithEnv(map[string]string{
						"PORT":                 "8080",
						"SERVICE_BINDING_ROOT": "/bindings",
					}).
					WithVolumes(fmt.Sprintf("%s/binding:/bindings/ca-certificates", source)).
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

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
				Expect(string(content)).To(ContainSubstring("<body>Hello World!</body>"))

				logs, err = docker.Container.Logs.Execute(container.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(logs).To(ContainLines(ContainSubstring("Added 1 additional CA certificate(s) to system truststore")))
			})
		})
	})
}
