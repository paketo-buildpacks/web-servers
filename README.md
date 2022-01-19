# Web Servers Paketo Buildpack

## `gcr.io/paketo-community/web-servers`

The Web Servers Paketo Buildpack provides a set of collaborating buildpacks that
enable the use of web servers. These buildpacks include:
- [NGINX CNB](https://github.com/paketo-buildpacks/nginx)
- [Apache HTTPD CNB](https://github.com/paketo-buildpacks/httpd)

The buildpack supports building applications that leverage NGINX or HTTPD web
servers. Usage examples can be found in the
[`samples`
repository](https://github.com/paketo-buildpacks/samples) under
the [`nginx`
directory](https://github.com/paketo-buildpacks/samples/tree/main/nginx) and
the [`httpd` directory
](https://github.com/paketo-buildpacks/samples/tree/main/httpd).

#### The Web Servers buildpack is only compatible with the following builder:
- [Paketo Full Builder](https://github.com/paketo-buildpacks/full-builder) (NGINX and HTTPD)
- [Paketo Base Builder](https://github.com/paketo-buildpacks/base-builder) (NGINX only)
**Note** that only HTTPD workloads are compatible with the Full Builder ONLY.

Check out the [Web Servers Paketo Buildpack docs](https://paketo.io/docs/howto/web-servers/) for more information.
