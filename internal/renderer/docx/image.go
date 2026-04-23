package docx

import (
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/types"
)

// EMU (English Metric Unit) conversion factors for OOXML drawing dimensions.
const (
	emuPerInch           = 914400
	emuPerCm             = 360000
	emuPerMm             = 36000
	emuPerPx             = 9525    // at 96 DPI
	defaultImageWidthEMU = 3657600 // 4 inches
)

func (r *docxRenderer) renderImageBlock(img *types.ImageBlock) error {
	src := r.resolveImagePath(img.Location)
	para := r.startParagraph(paragraphOptions{})
	if err := r.writeImageDrawing(para, src, img.Attributes); err != nil {
		r.writeTextRun(para, "[image: "+src+"]", runStyle{italic: true})
	}
	r.endParagraph(para)

	title, _ := img.Attributes.GetAsString(types.AttrTitle)
	if title != "" {
		number := r.ctx.GetAndIncrementImageCounter()
		captionPrefix, found := r.ctx.attributes.GetAsString(types.AttrFigureCaption)
		if !found || captionPrefix == "" {
			captionPrefix = "Figure"
		}
		if err := r.renderTextParagraph(captionPrefix+" "+strconv.Itoa(number)+". "+title, paragraphOptions{style: "Caption"}); err != nil {
			return err
		}
	}

	return nil
}

func (r *docxRenderer) renderInlineImage(para *strings.Builder, img *types.InlineImage) error {
	src := r.resolveImagePath(img.Location)
	if err := r.writeImageDrawing(para, src, img.Attributes); err != nil {
		r.writeTextRun(para, imageAlt(img.Attributes, src), runStyle{italic: true})
	}
	return nil
}

func (r *docxRenderer) resolveImagePath(location *types.Location) string {
	if location == nil {
		return ""
	}
	src := location.ToString()
	if imagesdir, found := r.ctx.attributes.GetAsString(types.AttrImagesDir); found {
		if !filepath.IsAbs(src) && !strings.Contains(src, "://") {
			src = filepath.Join(imagesdir, src)
		}
	}

	if !filepath.IsAbs(src) && !strings.Contains(src, "://") {
		dir := filepath.Dir(r.ctx.config.Filename)
		src = filepath.Join(dir, src)
	}
	return src
}

func (r *docxRenderer) writeImageDrawing(para *strings.Builder, src string, attrs types.Attributes) error {
	if src == "" || strings.Contains(src, "://") {
		return os.ErrNotExist
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	rID, name := r.doc.addImage(data, src)
	alt := imageAlt(attrs, src)
	width, height := imageSize(attrs)
	drawingID := r.doc.nextDrawingID()
	para.WriteString(`<w:r><w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0"><wp:extent cx="`)
	para.WriteString(strconv.FormatInt(width, 10))
	para.WriteString(`" cy="`)
	para.WriteString(strconv.FormatInt(height, 10))
	para.WriteString(`"/><wp:docPr id="`)
	para.WriteString(strconv.Itoa(drawingID))
	para.WriteString(`" name="`)
	para.WriteString(xmlAttr(name))
	para.WriteString(`" descr="`)
	para.WriteString(xmlAttr(alt))
	para.WriteString(`"/><wp:cNvGraphicFramePr><a:graphicFrameLocks noChangeAspect="1"/></wp:cNvGraphicFramePr><a:graphic><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:pic><pic:nvPicPr><pic:cNvPr id="`)
	para.WriteString(strconv.Itoa(drawingID))
	para.WriteString(`" name="`)
	para.WriteString(xmlAttr(name))
	para.WriteString(`"/><pic:cNvPicPr/></pic:nvPicPr><pic:blipFill><a:blip r:embed="`)
	para.WriteString(xmlAttr(rID))
	para.WriteString(`"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill><pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="`)
	para.WriteString(strconv.FormatInt(width, 10))
	para.WriteString(`" cy="`)
	para.WriteString(strconv.FormatInt(height, 10))
	para.WriteString(`"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr></pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r>`)
	return nil
}

func imageSize(attrs types.Attributes) (int64, int64) {
	width := dimensionToEMU(attrs.GetAsStringWithDefault(types.AttrWidth, ""))
	height := dimensionToEMU(attrs.GetAsStringWithDefault(types.AttrHeight, ""))
	if width == 0 {
		width = defaultImageWidthEMU
	}
	if height == 0 {
		height = width * 3 / 4
	}
	return width, height
}

func dimensionToEMU(value string) int64 {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasSuffix(value, "%") {
		return 0
	}
	unit := ""
	for _, suffix := range []string{"px", "in", "cm", "mm"} {
		if strings.HasSuffix(value, suffix) {
			unit = suffix
			value = strings.TrimSuffix(value, suffix)
			break
		}
	}
	n, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || n <= 0 {
		return 0
	}
	switch unit {
	case "in":
		return int64(n * emuPerInch)
	case "cm":
		return int64(n * emuPerCm)
	case "mm":
		return int64(n * emuPerMm)
	default:
		return int64(n * emuPerPx)
	}
}

func imageAlt(attrs types.Attributes, src string) string {
	if alt := attrs.GetAsStringWithDefault(types.AttrImageAlt, ""); alt != "" {
		return alt
	}
	u, err := url.Parse(src)
	if err != nil {
		return "[image: " + src + "]"
	}
	name := strings.TrimSuffix(filepath.Base(u.Path), filepath.Ext(u.Path))
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	if name == "" {
		return "[image: " + src + "]"
	}
	return name
}
