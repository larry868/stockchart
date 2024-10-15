# stockchart wasm example

stockchart wasm example is a simple webpage embedding the HTML5 web assembly stockchart component written in GO.

See more about the [stockchart go package](https://github.com/larry868/stockchart).

![snapshot](../snapshot.png)

## Runing the example

The example folder provides a webserver embedding the interactive stockchart if you want to run it by yourself. Install and launch [liveServer](https://marketplace.visualstudio.com/items?itemName=ritwickdey.LiveServer), then start-it or open http://localhost:5510/. You may need to tune up [settings.json](./.vscode/settings.json) to run liveserver in your environment.

You can also see it https://larry868.github.io/stockchart/

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

If you want to rebuild the wasm code, run the VSCODE task `buil wasm`.

### Using Web Assembly with go

Some documentation available here https://tinygo.org/docs/guides/webassembly/ and here https://github.com/golang/go/wiki/WebAssembly

Go provides a specific js file called `wasm_exec.js` that need to be served by your webpapp. This file is located in the ``/misc/wasm/`` subdirectory of your go root path. Usually we copy it to the folder containing all static files of your webapp, like `cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./examples/web` for the above example. It's important to get the version corresponding to your go environment, it's why we recommend to copy it from your GOROOT.

