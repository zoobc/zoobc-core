### ZooBC-Core

### Swagger

- install swagger https://www.npmjs.com/package/swagger with `npm install -g swagger`
- pull newest `schema` and run `./compile-go.sh` to recompile the go file and produce swagger definition from it.
- go to `/swagger`
- run `swagger mixin model/*.json service/*.json -o master.json` to combine the swagger definition files into single file.
- finally serve the swagger by `swagger serve master.json`
