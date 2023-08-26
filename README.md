## zserv

Simple HTTP server that serves files directly from .zip files without unpacking the zip file to the filesystem.

My usecase is to save locally mirrored websites (eg: using `wget`) as zip and then serve pages directly from that file. This prevents having to store the website as lot of small files - which cause clutter during copy / move / backup operations. Saving as ZIP files also saves some space.

Can specify custom port, host and website root relative to ZIP file. Host defaults to localhost (`127.0.0.1`) and will not be accessible from other devices on the network.

