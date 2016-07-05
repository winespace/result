package resulthandler

import (
	"encoding/csv"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"os"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type WebContext interface {
	ResponseWriter() http.ResponseWriter
	Request() *http.Request
}

type Result interface {
	Do(WebContext) error
}

type ResultHead struct {
	Code int
}

func (r *ResultHead) Do(ctx WebContext) error {
	ctx.ResponseWriter().WriteHeader(r.Code)
	return nil
}

type ResultText struct {
	Text string
	Code int
}

func (r *ResultText) Do(ctx WebContext) error {
	w := ctx.ResponseWriter()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(r.Code)
	_, err := io.WriteString(w, r.Text)
	return err
}

type ResultHTML struct {
	Text string
	Name string
	*template.Template
	Data interface{}
}

func (r *ResultHTML) Do(ctx WebContext) error {
	w := ctx.ResponseWriter()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Text != "" {
		_, err := io.WriteString(w, r.Text)
		return err
	}
	if r.Name != "" {
		return r.Template.ExecuteTemplate(
			ctx.ResponseWriter(),
			r.Name,
			r.Data,
		)
	}
	return r.Template.Execute(ctx.ResponseWriter(), r.Data)
}

type ResultJSON struct {
	Data interface{}
	Code int
}

func (r *ResultJSON) Do(ctx WebContext) error {
	jsonStr, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}
	if r.Code == 0 {
		r.Code = 200
	}
	w := ctx.ResponseWriter()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Code)
	_, err = w.Write(jsonStr)
	return err
}

type ResultCSV struct {
	Data [][]string
	Attachment bool
	Code int
}

func (r *ResultCSV) Do(ctx WebContext) error {
	w := ctx.ResponseWriter()
	w.Header().Set("Content-Type", "text/csv; char=utf-8")
	if r.Attachment {
		w.Header().Set("Content-Disposition", "attachment")
	}
	w.WriteHeader(r.Code)
	csvw := csv.NewWriter(transform.NewWriter(w, japanese.ShiftJIS.NewEncoder()))
	return csvw.WriteAll(r.Data)
}

type ResultFile struct {
	Path string
}

func (r *ResultFile) Do(ctx WebContext) error {
	if _, err := os.Stat(r.Path); err != nil {
		return err
	}
	http.ServeFile(ctx.ResponseWriter(), ctx.Request(), r.Path)
	return nil
}

type ResultData struct {
	Data        []byte
	ContentType string
}

func (r *ResultData) Do(ctx WebContext) error {
	w := ctx.ResponseWriter()
	if r.ContentType != "" {
		w.Header().Set("Content-Type", r.ContentType)
	}
	_, err := w.Write(r.Data)
	return err
}

type ResultRedirect struct {
	URL  string
	Code int
}

func (r *ResultRedirect) Do(ctx WebContext) error {
	if r.Code == 0 {
		r.Code = 303
	}
	http.Redirect(
		ctx.ResponseWriter(),
		ctx.Request(),
		r.URL,
		r.Code,
	)
	return nil
}
