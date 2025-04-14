package compiler

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const networkTemplate = `
// #include <stdlib.h>
// #include <curl/curl.h>
//
// typedef struct {
//     char* url;
//     char* method;
//     char* data;
// } Request;
//
// void perform_request(Request* req) {
//     CURL *curl = curl_easy_init();
//     if(curl) {
//         curl_easy_setopt(curl, CURLOPT_URL, req->url);
//         curl_easy_setopt(curl, CURLOPT_CUSTOMREQUEST, req->method);
//         if(req->data) {
//             curl_easy_setopt(curl, CURLOPT_POSTFIELDS, req->data);
//         }
//         curl_easy_perform(curl);
//         curl_easy_cleanup(curl);
//     }
// }
import "C"
import (
    "unsafe"
    "sync"
)

`

type CodeGenerator struct {
	ast *AST
}

func (c *CodeGenerator) Generate() (string, error) {
	code := "package main\n\nimport (\n\t\"fmt\"\n"

	// Add imports
	for _, imp := range c.ast.Imports {
		code += fmt.Sprintf("\t. \"%s\"\n", imp)
	}
	code += ")\n\n"

	// Generate functions
	for _, stmt := range c.ast.Statements {
		switch s := stmt.(type) {
		case *ArrayDeclaration:
			code += fmt.Sprintf("var %s = []interface{}{", s.Name)
			for i, elem := range s.Elements {
				if i > 0 {
					code += ", "
				}
				switch v := elem.(type) {
				case string:
					if _, err := strconv.Atoi(v); err == nil {
						code += v // Unquoted number string
					} else {
						code += fmt.Sprintf("%q", v)
					}
				case int:
					code += fmt.Sprintf("%d", v)
				case float64:
					code += fmt.Sprintf("%f", v)

				case *NetworkCall:
					if networkCall, ok := elem.(*NetworkCall); ok {
						code += c.generateNetworkCall(networkCall)
					} else {
						code += fmt.Sprintf("%v", elem)
					}
				case *ParallelBlock:
					if parallelBlock, ok := elem.(*ParallelBlock); ok {
						code += c.generateParallelBlock(parallelBlock)
					} else {
						// Handle the case where elem is not a *ParallelBlock
						code += fmt.Sprintf("%v", elem)
					}
				}
			}
			code += "}\n"
		case *FunctionDecl:
			code += fmt.Sprintf("func %s(%s) interface{} {\n",
				s.Name, strings.Join(s.Params, ", "))
			code += "\t// Auto-generated function\n"
			code += "}\n\n"
		case *VarStatement:
			code += fmt.Sprintf("var %s = %s\n", s.Name, s.Value)
		case *DoStatement:
			code += "func main() {\n"
			code += fmt.Sprintf("\tfmt.Println(%s)\n", strings.Join(toStrings(s.Args), " "))
			code += "}\n"
		}
	}

	return code, nil
}

func toStrings(args []interface{}) []string {
	var res []string
	for _, a := range args {
		res = append(res, fmt.Sprint(a))
	}
	return res
}

func CompileToGo(ast *AST, outputPath string) error {
	gen := &CodeGenerator{ast: ast}
	code, err := gen.Generate()
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, []byte(code), 0644)
}

func (c *CodeGenerator) generateNetworkCall(call *NetworkCall) string {
	if call.Parallel {
		return fmt.Sprintf(`
            go func() {
                req := C.Request{
                    url: C.CString("%s"),
                    method: C.CString("%s"),
                    data: C.CString("%s"),
                }
                C.perform_request(&req)
                defer C.free(unsafe.Pointer(req.url))
                defer C.free(unsafe.Pointer(req.method))
                defer C.free(unsafe.Pointer(req.data))
            }()
        `, call.URL, call.Method, call.Payload)
	}

	return fmt.Sprintf(`
        req := C.Request{
            url: C.CString("%s"),
            method: C.CString("%s"),
            data: C.CString("%s"),
        }
        C.perform_request(&req)
        defer C.free(unsafe.Pointer(req.url))
        defer C.free(unsafe.Pointer(req.method))
        defer C.free(unsafe.Pointer(req.data))
    `, call.URL, call.Method, call.Payload)
}

func (c *CodeGenerator) generateParallelBlock(block *ParallelBlock) string {
	code := `
        var wg sync.WaitGroup
        sem := make(chan struct{}, 1000) // Limit concurrency
    `

	for _, call := range block.Calls {
		code += fmt.Sprintf(`
            wg.Add(1)
            go func() {
                defer wg.Done()
                sem <- struct{}{}
                defer func() { <-sem }()
                req := C.Request{
                    url: C.CString("%s"),
                    method: C.CString("%s"),
                    data: C.CString("%s"),
                }
                C.perform_request(&req)
                defer C.free(unsafe.Pointer(req.url))
                defer C.free(unsafe.Pointer(req.method))
                defer C.free(unsafe.Pointer(req.data))
            }()
        `, call.URL, call.Method, call.Payload)
	}

	code += `
        wg.Wait()
    `
	return code
}
