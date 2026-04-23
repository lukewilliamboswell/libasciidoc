package sgml

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

func (r *sgmlRenderer) renderDocumentDetails(ctx *context) (string, error) {
	if !ctx.attributes.Has(types.AttrAuthors) {
		log.Debugf("no authors to render")
		return "", nil
	}
	authors, err := r.renderDocumentAuthorsDetails(ctx)
	if err != nil {
		return "", fmt.Errorf("error while rendering the document details: %w", err)
	}
	documentDetailsBuff := &bytes.Buffer{}
	revLabel, _ := ctx.attributes.GetAsString(types.AttrVersionLabel)
	revNumber, _ := ctx.attributes.GetAsString("revnumber")
	revDate, _ := ctx.attributes.GetAsString("revdate")
	revRemark, _ := ctx.attributes.GetAsString("revremark")
	tmpl, err := r.documentDetails()
	if err != nil {
		return "", fmt.Errorf("unable to load document details template: %w", err)
	}
	if err = tmpl.Execute(documentDetailsBuff, struct {
		Authors   string
		RevLabel  string
		RevNumber string
		RevDate   string
		RevRemark string
	}{
		Authors:   authors,
		RevLabel:  revLabel,
		RevNumber: revNumber,
		RevDate:   revDate,
		RevRemark: revRemark,
	}); err != nil {
		return "", fmt.Errorf("error while rendering the document details: %w", err)
	}
	return documentDetailsBuff.String(), nil
}

func (r *sgmlRenderer) renderDocumentAuthorsDetails(ctx *context) (string, error) {
	authorsDetailsBuff := &strings.Builder{}
	i := 1
	for {
		var authorKey string
		var emailKey string
		var index string
		if i == 1 {
			authorKey = "author"
			emailKey = "email"
			index = ""
		} else {
			index = strconv.Itoa(i)
			authorKey = "author_" + index
			emailKey = "email_" + index
		}
		// having at least one author is the minimal requirement for document details
		if author, ok := ctx.attributes.GetAsString(authorKey); ok {
			if i > 1 {
				authorsDetailsBuff.WriteString("\n")
			}
			email, _ := ctx.attributes.GetAsString(emailKey)
			tmpl, err := r.documentAuthorDetails()
			if err != nil {
				return "", fmt.Errorf("unable to load document authors template: %w", err)
			}
			if err := tmpl.Execute(authorsDetailsBuff, struct {
				Index string
				Name  string
				Email string
			}{
				Index: index,
				Name:  author,
				Email: email,
			}); err != nil {
				return "", fmt.Errorf("error while rendering the document authors: %w", err)
			}
			// if there were authors before, need to insert a `\n`
			i++
		} else {
			break
		}
	}
	return authorsDetailsBuff.String(), nil
}
