// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package apigen

import (
	"fmt"
	"go/format"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/zeebo/errs"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"storj.io/common/uuid"
)

// MustWriteGo writes generated Go code into a file.
func (a *API) MustWriteGo(path string) {
	generated, err := a.generateGo()
	if err != nil {
		panic(errs.Wrap(err))
	}

	err = os.WriteFile(path, generated, 0644)
	if err != nil {
		panic(errs.Wrap(err))
	}
}

// generateGo generates api code and returns an output.
func (a *API) generateGo() ([]byte, error) {
	var result string

	p := func(format string, a ...interface{}) {
		result += fmt.Sprintf(format+"\n", a...)
	}

	getPackageName := func(path string) string {
		pathPackages := strings.Split(path, "/")
		return pathPackages[len(pathPackages)-1]
	}

	p("// AUTOGENERATED BY private/apigen")
	p("// DO NOT EDIT.")
	p("")

	p("package %s", a.PackageName)
	p("")

	p("import (")
	p(`"context"`)
	p(`"encoding/json"`)
	p(`"net/http"`)
	p(`"time"`)
	p("")
	p(`"github.com/gorilla/mux"`)
	p(`"github.com/zeebo/errs"`)
	p(`"go.uber.org/zap"`)
	p("")

	p(`"storj.io/common/uuid"`)
	p(`"storj.io/storj/private/api"`)

	for _, group := range a.EndpointGroups {
		for _, method := range group.Endpoints {
			if method.Request != nil {
				path := reflect.TypeOf(method.Request).Elem().PkgPath()
				pn := getPackageName(path)
				if pn == a.PackageName {
					continue
				}

				p(`"%s"`, path)
			}
			if method.Response != nil {
				path := reflect.TypeOf(method.Response).Elem().PkgPath()
				pn := getPackageName(path)
				if pn == a.PackageName {
					continue
				}

				p(`"%s"`, path)
			}
		}
	}

	p(")")
	p("")

	p("const dateLayout = \"2006-01-02T15:04:05.000Z\"")
	p("")

	for _, group := range a.EndpointGroups {
		p("var Err%sAPI = errs.Class(\"%s %s api\")", cases.Title(language.Und).String(group.Prefix), a.PackageName, group.Prefix)
	}

	p("")

	for _, group := range a.EndpointGroups {
		p("type %sService interface {", group.Name)
		for _, e := range group.Endpoints {
			var params string
			for _, param := range e.Params {
				params += param.Type.String() + ", "
			}

			if e.Response != nil {
				responseType := reflect.TypeOf(e.Response)
				p("%s(context.Context, "+params+") (%s, api.HTTPError)", e.MethodName, a.handleTypesPackage(responseType))
			} else {
				p("%s(context.Context, "+params+") (api.HTTPError)", e.MethodName)
			}
		}
		p("}")
		p("")
	}

	for _, group := range a.EndpointGroups {
		p("// %sHandler is an api handler that exposes all %s related functionality.", group.Name, group.Prefix)
		p("type %sHandler struct {", group.Name)
		p("log *zap.Logger")
		p("service %sService", group.Name)
		p("auth api.Auth")
		p("}")
		p("")
	}

	for _, group := range a.EndpointGroups {
		p(
			"func New%s(log *zap.Logger, service %sService, router *mux.Router, auth api.Auth) *%sHandler {",
			group.Name,
			group.Name,
			group.Name,
		)
		p("handler := &%sHandler{", group.Name)
		p("log: log,")
		p("service: service,")
		p("auth: auth,")
		p("}")
		p("")
		p("%sRouter := router.PathPrefix(\"/api/v0/%s\").Subrouter()", group.Prefix, group.Prefix)
		for pathMethod, endpoint := range group.Endpoints {
			handlerName := "handle" + endpoint.MethodName
			p("%sRouter.HandleFunc(\"%s\", handler.%s).Methods(\"%s\")", group.Prefix, pathMethod.Path, handlerName, pathMethod.Method)
		}
		p("")
		p("return handler")
		p("}")
		p("")
	}

	for _, group := range a.EndpointGroups {
		for pathMethod, endpoint := range group.Endpoints {
			p("")
			handlerName := "handle" + endpoint.MethodName
			p("func (h *%sHandler) %s(w http.ResponseWriter, r *http.Request) {", group.Name, handlerName)
			p("ctx := r.Context()")
			p("var err error")
			p("defer mon.Task()(&ctx)(&err)")
			p("")

			p("w.Header().Set(\"Content-Type\", \"application/json\")")
			p("")

			if !endpoint.NoCookieAuth || !endpoint.NoAPIAuth {
				if !endpoint.NoCookieAuth && !endpoint.NoAPIAuth {
					p("ctx, err = h.auth.IsAuthenticated(ctx, r, true, true)")
				}
				if endpoint.NoCookieAuth && !endpoint.NoAPIAuth {
					p("ctx, err = h.auth.IsAuthenticated(ctx, r, false, true)")
				}
				if !endpoint.NoCookieAuth && endpoint.NoAPIAuth {
					p("ctx, err = h.auth.IsAuthenticated(ctx, r, true, false)")
				}
				p("if err != nil {")
				p("api.ServeError(h.log, w, http.StatusUnauthorized, err)")
				p("return")
				p("}")
				p("")
			}

			switch pathMethod.Method {
			case http.MethodGet:
				for _, param := range endpoint.Params {
					switch param.Type {
					case reflect.TypeOf(uuid.UUID{}):
						handleUUIDQuery(p, param)
						continue
					case reflect.TypeOf(time.Time{}):
						handleTimeQuery(p, param)
						continue
					case reflect.TypeOf(""):
						handleStringQuery(p, param)
						continue
					}
				}
			case http.MethodPatch:
				for _, param := range endpoint.Params {
					if param.Type == reflect.TypeOf(uuid.UUID{}) {
						handleUUIDParam(p, param)
					} else {
						handleBody(p, param)
					}
				}
			case http.MethodPost:
				for _, param := range endpoint.Params {
					handleBody(p, param)
				}
			case http.MethodDelete:
				for _, param := range endpoint.Params {
					handleUUIDParam(p, param)
				}
			}

			var methodFormat string
			if endpoint.Response != nil {
				methodFormat = "retVal, httpErr := h.service.%s(ctx, "
			} else {
				methodFormat = "httpErr := h.service.%s(ctx, "
			}

			switch pathMethod.Method {
			case http.MethodGet:
				for _, methodParam := range endpoint.Params {
					methodFormat += methodParam.Name + ", "
				}
			case http.MethodPatch:
				for _, methodParam := range endpoint.Params {
					if methodParam.Type == reflect.TypeOf(uuid.UUID{}) {
						methodFormat += methodParam.Name + ", "
					} else {
						methodFormat += "*" + methodParam.Name + ", "
					}
				}
			case http.MethodPost:
				for _, methodParam := range endpoint.Params {
					methodFormat += "*" + methodParam.Name + ", "
				}
			case http.MethodDelete:
				for _, methodParam := range endpoint.Params {
					methodFormat += methodParam.Name + ", "
				}
			}

			methodFormat += ")"
			p(methodFormat, endpoint.MethodName)
			p("if httpErr.Err != nil {")
			p("api.ServeError(h.log, w, httpErr.Status, httpErr.Err)")
			if endpoint.Response == nil {
				p("}")
				p("}")
				continue
			}
			p("return")
			p("}")

			p("")
			p("err = json.NewEncoder(w).Encode(retVal)")
			p("if err != nil {")
			p("h.log.Debug(\"failed to write json %s response\", zap.Error(Err%sAPI.Wrap(err)))", endpoint.MethodName, cases.Title(language.Und).String(group.Prefix))
			p("}")
			p("}")
		}
	}

	output, err := format.Source([]byte(result))
	if err != nil {
		return nil, err
	}

	return output, nil
}

// handleTypesPackage handles the way some type is used in generated code.
// If type is from the same package then we use only type's name.
// If type is from external package then we use type along with its appropriate package name.
func (a *API) handleTypesPackage(t reflect.Type) interface{} {
	if strings.HasPrefix(t.String(), a.PackageName) {
		return t.Elem().Name()
	}

	return t
}

// handleStringQuery handles request query param of type string.
func handleStringQuery(p func(format string, a ...interface{}), param Param) {
	p("%s := r.URL.Query().Get(\"%s\")", param.Name, param.Name)
	p("if %s == \"\" {", param.Name)
	p("api.ServeError(h.log, w, http.StatusBadRequest, errs.New(\"parameter '%s' can't be empty\"))", param.Name)
	p("return")
	p("}")
	p("")
}

// handleUUIDQuery handles request query param of type uuid.UUID.
func handleUUIDQuery(p func(format string, a ...interface{}), param Param) {
	p("%s, err := uuid.FromString(r.URL.Query().Get(\"%s\"))", param.Name, param.Name)
	p("if err != nil {")
	p("api.ServeError(h.log, w, http.StatusBadRequest, err)")
	p("return")
	p("}")
	p("")
}

// handleTimeQuery handles request query param of type time.Time.
func handleTimeQuery(p func(format string, a ...interface{}), param Param) {
	p("%s, err := time.Parse(dateLayout, r.URL.Query().Get(\"%s\"))", param.Name, param.Name)
	p("if err != nil {")
	p("api.ServeError(h.log, w, http.StatusBadRequest, err)")
	p("return")
	p("}")
	p("")
}

// handleUUIDParam handles request inline param of type uuid.UUID.
func handleUUIDParam(p func(format string, a ...interface{}), param Param) {
	p("%sParam, ok := mux.Vars(r)[\"%s\"]", param.Name, param.Name)
	p("if !ok {")
	p("api.ServeError(h.log, w, http.StatusBadRequest, errs.New(\"missing %s route param\"))", param.Name)
	p("return")
	p("}")
	p("")

	p("%s, err := uuid.FromString(%sParam)", param.Name, param.Name)
	p("if err != nil {")
	p("api.ServeError(h.log, w, http.StatusBadRequest, err)")
	p("return")
	p("}")
	p("")
}

// handleBody handles request body.
func handleBody(p func(format string, a ...interface{}), param Param) {
	p("%s := &%s{}", param.Name, param.Type)
	p("if err = json.NewDecoder(r.Body).Decode(&%s); err != nil {", param.Name)
	p("api.ServeError(h.log, w, http.StatusBadRequest, err)")
	p("return")
	p("}")
	p("")
}
