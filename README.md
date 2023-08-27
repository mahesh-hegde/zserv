## zserv

Simple HTTP server that serves files directly from .zip files without unpacking the zip file to the filesystem.

My usecase is to save locally mirrored websites (eg: using `wget`) as zip and then serve pages directly from that file. This prevents having to store the website as lot of small files - which cause clutter during copy / move / backup operations. Saving as ZIP files also saves some space.

Can specify custom port, host and website root relative to ZIP file. Host defaults to localhost (`127.0.0.1`) and will not be accessible from other devices on the network.

## Install

### Go
```bash
go install github.com/mahesh-hegde/zserv@v0.1.0
```

### Docker / Podman
`zserv` is available on [ghcr](https://ghcr.io/mahesh-hegde/zserv)

There's one caveat with `docker` / `podman` - the container is a separate host attached with a bridge network. So the default binding to 127.0.0.1 will not be accessible from the host computer.

Either use `--network host` option, or provide `--host 0.0.0.0` to `zserv` and restrict the docker port forwarding itself. 

```bash
docker run -dp "127.0.0.1:8088:8088" -v $PWD:/app/ ghcr.io/mahesh-hegde/zserv:v0.1.0 --host '0.0.0.0' my_saved_website.zip

## or simply attach to host network
docker run -dp 8088:8088 -v $PWD:/app --network host ghcr.io/mahesh-hegde/zserv my_saved_website.zip

## use -tp instead of -dp if you want to see the console output
```
