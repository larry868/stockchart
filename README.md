# gowebstockchart

HTML5 stock chart module in go 

Generate interactive, responsive, and perfomant chart with HTML5 canvas and web assembly, and embed it into any HTML pages. 

![snapshot](doc/snapshot.png)

## features

- draw candlestick with OHLC series
- X time axis navigator
- X time axis with auto scale and autolabelling
- zoom-in and zoom-out with the mouse wheel inside the navbar
- shift left and right with Shift-Key and the mouse wheel inside the navbar
- Y value axis, with auto scale and auto labelling
- responsive: handle resize event, and browser zoom
- embedding chart with a single <stockchart> HTML elemnt

## characteristics

- fast drawing on any browser accepting HTML5 canvas & Webassembly
- only GO, no JS

## Runing the example

The example folder provides a webserver example embedding an interactive gowebstockchart.

The file structure is the followwing one
```
example/
+- wasm/            
 - dataset.go      a simgle function to build the dataset for the example
 - main.go         source code of the wasm code to be built to be loaded by the browser
+- webapp/          
 - index.html      static html file of this example
 - main.wasm       compiled version of the code to be loaded by the browser
 - myapp.js        a minimal and required js code to be loaded by your web pages, to be customized
 - wasm_exec.js    provided by go, see here after
+- webserver
 - main.go         an optional simple webserver serving static files located in webapp, see here after
```

## Technicals

Written in go 1.19

The Web Assembly code generated is based on the [webapi package](https://github.com/gowebapi/webapi).

### Use Web Assembly with go

Some documentation available here https://tinygo.org/docs/guides/webassembly/ and here https://github.com/golang/go/wiki/WebAssembly

Go provides a specific js file called `wasm_exec.js` that need to be served by your webpapp. This file is located in the ``/misc/wasm/`` subdirectory of your go root path. Usually we copy it to the folder containing all static files of your webapp, like `cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" /example/webapp` for the above example. It's important to get the version corresponding to your go environment, it's why we recomend to copy it from your GOROOT.

## Change log

- v0.1.0 alpha: contains lines commented with `// DEBUG:` for debug purpose only
