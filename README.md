# Web Servers Paketo Buildpack

## `paketo-buildpacks/web-servers`

The Web Servers Paketo Buildpack provides a set of collaborating buildpacks that
enable the use of web servers. These buildpacks include:
- [NGINX CNB](https://github.com/paketo-buildpacks/nginx)
- [Apache HTTPD CNB](https://github.com/paketo-buildpacks/httpd)
- [Node Engine CNB](https://github.com/paketo-buildpacks/node-engine)
- [Yarn CNB](https://github.com/paketo-buildpacks/yarn)
- [Yarn Install CNB](https://github.com/paketo-buildpacks/yarn-install)
- [NPM Install CNB](https://github.com/paketo-buildpacks/npm-install)

The buildpack supports building applications that leverage NGINX or HTTPD web
servers as well as JavaScript Frontend apps. Usage examples can be found in the
[`samples`
repository](https://github.com/paketo-buildpacks/samples) under
the [`web-servers`
directory](https://github.com/paketo-buildpacks/samples/tree/main/web-servers).

#### The Web Servers buildpack is only compatible with the following builder:
- [Paketo Full Builder](https://github.com/paketo-buildpacks/full-builder) (NGINX and HTTPD)
- [Paketo Base Builder](https://github.com/paketo-buildpacks/base-builder) (NGINX only)
**Note** that HTTPD workloads are compatible with the Full Builder ONLY.

This buildpack also includes the following utility buildpacks:
- [Procfile CNB](https://github.com/paketo-buildpacks/procfile)
- [Environment Variables CNB](https://github.com/paketo-buildpacks/environment-variables)
- [Image Labels CNB](https://github.com/paketo-buildpacks/image-labels)
- [CA Certificates CNB](https://github.com/paketo-buildpacks/ca-certificates)
- [Node Run Script CNB](https://github.com/paketo-buildpacks/node-run-script)

Check out the [Web Servers Paketo Buildpack docs](https://paketo.io/docs/howto/web-servers/) for more information.
