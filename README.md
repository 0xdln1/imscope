# imscope

Get scope from immunefi


### Usage:

```
Usage: go run main.go <assetType>
Available options for <assetType>:
  - smart_contract
  - websites_and_applications
  - blockchain_dlt
  - all (fetches all asset types)
```

Example: fetch only assets which are of type `websites_and_applications`

```
go run imscope.go websites_and_applications | tee output.txt
```

### Installation:

```
go install github.com/0xdln1/imscope@latest
```
