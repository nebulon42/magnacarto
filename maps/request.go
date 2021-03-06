package maps

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type BBOX struct {
	MinX, MinY, MaxX, MaxY float64
}

type Request struct {
	HTTP     *http.Request
	Query    url.Values
	Width    int
	Height   int
	BBOX     BBOX
	EPSGCode int
	Format   string
}

type MissingParamError struct {
	Param string
}

func (e *MissingParamError) Error() string {
	return fmt.Sprintf("Missing parameter '%s'", e.Param)
}

type InvalidParamError struct {
	Param string
	Value string
}

func (e *InvalidParamError) Error() string {
	return fmt.Sprintf("Invalid parameter '%s=%s'", e.Param, e.Value)
}

func parseQueryUpper(q string) (url.Values, error) {
	origQuery, err := url.ParseQuery(q)
	if err != nil {
		return nil, err
	}

	query := make(url.Values)
	for k, vs := range origQuery {
		for _, v := range vs {
			query.Add(strings.ToUpper(k), v)
		}
	}
	return query, nil
}

func ParseMapRequest(r *http.Request) (*Request, error) {
	query, err := parseQueryUpper(r.URL.RawQuery)
	if err != nil {
		return nil, err
	}

	req := &Request{
		HTTP:     r,
		Query:    query,
		EPSGCode: 3857,
	}

	req.Width, req.Height, err = parseSize(req.Query)
	if err != nil {
		return nil, err
	}
	req.BBOX, err = parseBBOX(req.Query)
	if err != nil {
		return nil, err
	}

	req.Format, err = parseFormat(req.Query)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func parseSize(q url.Values) (width, height int, err error) {

	w := q.Get("WIDTH")
	if w == "" {
		return width, height, &MissingParamError{"WIDTH"}
	}

	h := q.Get("HEIGHT")
	if h == "" {
		return width, height, &MissingParamError{"HEIGHT"}
	}
	var tmp uint64
	tmp, err = strconv.ParseUint(w, 10, 32)
	if err != nil || tmp == 0 {
		return width, height, &InvalidParamError{"WIDTH", w}
	}
	width = int(tmp)
	tmp, err = strconv.ParseUint(h, 10, 32)
	if err != nil || tmp == 0 {
		return width, height, &InvalidParamError{"HEIGHT", h}
	}
	height = int(tmp)
	return width, height, nil
}

func parseBBOX(q url.Values) (BBOX, error) {
	var err error
	var bbox BBOX

	bboxStr := q.Get("BBOX")
	if bboxStr == "" {
		return bbox, &MissingParamError{"BBOX"}
	}

	bboxParts := strings.Split(bboxStr, ",")
	if len(bboxParts) != 4 {
		return bbox, &InvalidParamError{"BBOX", bboxStr}
	}

	bbox.MinX, err = strconv.ParseFloat(bboxParts[0], 64)
	if err != nil {
		return bbox, &InvalidParamError{"BBOX", bboxStr}
	}
	bbox.MinY, err = strconv.ParseFloat(bboxParts[1], 64)
	if err != nil {
		return bbox, &InvalidParamError{"BBOX", bboxStr}
	}
	bbox.MaxX, err = strconv.ParseFloat(bboxParts[2], 64)
	if err != nil {
		return bbox, &InvalidParamError{"BBOX", bboxStr}
	}
	bbox.MaxY, err = strconv.ParseFloat(bboxParts[3], 64)
	if err != nil {
		return bbox, &InvalidParamError{"BBOX", bboxStr}
	}

	return bbox, nil
}

func parseFormat(q url.Values) (string, error) {
	format := q.Get("FORMAT")
	if format == "" {
		return "png256", nil
	}

	if strings.HasPrefix(format, "image/png") {
		if format == "image/png; mode=24bit" {
			return "png32", nil
		}
		return "png256", nil
	} else if strings.HasPrefix(format, "image/jpeg") {
		return "jpeg", nil
	} else {
		return "png256", &InvalidParamError{"FORMAT", format}
	}
}
