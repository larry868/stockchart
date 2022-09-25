# gowebstockchart

gowebstockchart is a simple webpage embedding the HTML5 web assembly stock chart written in GO.

See more about the [stockchart go package](https://github.com/sunraylab/stockchart).

![snapshot](snapshot.png)

## Runing the example

The example folder provides a webserver embedding the interactive stockchart.

The file structure is the followwing one
```text
example/
+- wasm/            
 - dataset.go      a simgle function to build the dataset for the example
 - main.go         source code of the wasm code to be built to be loaded by the browser
+- webapp/          
 - index.html      static html file of this example
 - main.wasm       compiled version of the code to be loaded by the browser
 - myapp.js        a minimal and required js code to be loaded by your web pages, to be customized
 - wasm_exec.js    provided by go, see here after
```

Install and launch [liveServer](https://marketplace.visualstudio.com/items?itemName=ritwickdey.LiveServer), then start-it or open http://localhost:5510/.

You may need to tune up [settings.json](./.vscode/settings.json) to run liveserver in your environment.

If you want to rebuild the wasm code, run the task `buil wasm`.

### Using Web Assembly with go

Some documentation available here https://tinygo.org/docs/guides/webassembly/ and here https://github.com/golang/go/wiki/WebAssembly

Go provides a specific js file called `wasm_exec.js` that need to be served by your webpapp. This file is located in the ``/misc/wasm/`` subdirectory of your go root path. Usually we copy it to the folder containing all static files of your webapp, like `cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" /example/webapp` for the above example. It's important to get the version corresponding to your go environment, it's why we recomend to copy it from your GOROOT.

