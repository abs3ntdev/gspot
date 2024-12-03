package listenbrainz

import (
	"bytes"
	"strconv"
)

type RadioPromptBuilder struct {
	px RadioParameters
}

func (o *RadioPromptBuilder) Add(name string, values ...string) *RadioPromptBuilder {
	o.px = append(o.px, RadioParameter{
		Name:   name,
		Values: values,
	})
	return o
}

func (o *RadioPromptBuilder) AddParameter(p RadioParameter) *RadioPromptBuilder {
	o.px = append(o.px, p)
	return o
}

func (o *RadioPromptBuilder) AddWithCount(name string, count int, values ...string) *RadioPromptBuilder {
	o.px = append(o.px, RadioParameter{
		Name:   name,
		Count:  count,
		Values: values,
	})
	return o
}
func (o *RadioPromptBuilder) AddWithOption(name string, option string, values ...string) *RadioPromptBuilder {
	o.px = append(o.px, RadioParameter{
		Name:   name,
		Option: option,
		Values: values,
	})
	return o
}

func (o *RadioPromptBuilder) String() string {
	val, _ := o.px.MarshalText()
	return string(val)
}

type RadioParameters []RadioParameter

func (r RadioParameters) MarshalText() ([]byte, error) {
	o := &bytes.Buffer{}
	for pidx, v := range r {
		o.WriteString(v.Name)
		o.WriteString(":(")
		for idx, vv := range v.Values {
			o.WriteString(vv)
			if len(v.Values) > 1 && idx != len(v.Values)-1 {
				o.WriteString(",")
			}
		}
		o.WriteString(")")
		if v.Count > 0 {
			o.WriteString(":" + strconv.Itoa(v.Count))
		}
		if v.Option != "" {
			o.WriteString(":" + v.Option)
		}
		if len(r) > 1 && pidx != len(r)-1 {
			o.WriteString(" ")
		}
	}
	return o.Bytes(), nil
}

type RadioParameter struct {
	Name   string   `json:"name"`
	Values []string `json:"value"`
	Count  int      `json:"count"`
	Option string   `json:"options"`
}
