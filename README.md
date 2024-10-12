# makeserver

`makeserver` is a small Go utility to generate a server-in-a-binary. Run it as
```
makeserver <input_folder> -o <output_binary>
```

It will create a standalone binary which serves your website on port 8000.

Currently it doesn't support any options, but I want to add support for:
  1. Specifying the port
  2. Support for HTTPS
  3. Options to control the URI to file mapping
