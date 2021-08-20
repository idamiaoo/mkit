package ucloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/beego/beego/v2/client/httplib"
	"github.com/spf13/cobra"
)

func formatCode(code []byte) []byte {
	formatted, err := format.Source(code)
	if err != nil {
		return code
	}
	return formatted
}

type Models struct {
	Models []*Model
}

type Model struct {
	Name        string     `json:"name"`
	Description string     `json:"desc"`
	Params      Parameters `json:"params"`
}

type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"desc"`
	ArrayType   string `json:"arrType"`
	Text        string
}

type Parameters []Parameter

func (ps *Parameters) String() string {
	return fmt.Sprintf("%#v", *ps)
}

func (ps *Parameters) Set(value string) error {
	rs := strings.Split(value, "$")
	if len(rs) < 2 {
		return fmt.Errorf("parameter is too short")
	}
	p := Parameter{}
	text := strings.Join(rs[0:2], "  ")
	if len(rs) > 2 {
		text += " `" + strings.Join(rs[2:], "  ") + "`"
	}
	p.Text = text
	*ps = append(*ps, p)
	return nil
}

func (ps *Parameters) Type() string {
	return "actionParameters"
}

type ActionDescribe struct {
	Name        string `json:"name"`
	Description string `json:"cName"`
	BaseRequest string
	Request     Parameters `json:"request,omitempty"`
	Response    Parameters `json:"response,omitempty"`
}

func (action ActionDescribe) String() string {
	bs, _ := json.Marshal(action)
	return string(bs)
}

func NewActionDescribe() *ActionDescribe {
	return &ActionDescribe{Request: make([]Parameter, 0, 1), Response: make([]Parameter, 0, 1)}
}

func parseParameter(p *Parameter) {
	switch p.Name {
	case "", "Region", "Zone", "ProjectId":
		return
	}
	text := p.Name + " "
	var tag []string
	switch p.Type {
	case "array":
		text += "[]" + p.ArrayType
	case "object":
		text += p.ArrayType
	case "float":
		text += "float64"
	default:
		text += p.Type
	}
	if p.Required {
		tag = append(tag, `validate:"required"`)
	}
	if p.Description != "" {
		tag = append(tag, fmt.Sprintf(`desc:"%s"`, p.Description))
	}
	if len(tag) > 0 {
		p.Text = fmt.Sprintf("%s `%s`", text, strings.Join(tag, " "))
	} else {
		p.Text = text
	}
}

func GenerateModels(currpath string, models []*Model) error {
	for _, v := range models {
		for i := range v.Params {
			parseParameter(&v.Params[i])
		}
	}
	t := template.Must(template.New("model").Parse(modelTpl))
	buff := bytes.NewBuffer(nil)
	if err := t.Execute(buff, &Models{Models: models}); err != nil {
		fmt.Printf("render template error %s\n", err)
		return err
	}
	code := formatCode(buff.Bytes())
	fp := path.Join(currpath, "controllers")
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		if err := os.MkdirAll(fp, 0777); err != nil {
			fmt.Printf("Could not create controllers directory: %s\n", err)
			return err
		}
	}
	fPath := path.Join(fp, "models.go")
	f, err := os.OpenFile(fPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	f.Write(code)
	return nil
}

func GenerateActionController(currpath string, action *ActionDescribe) error {
	if action.BaseRequest == "" {
		action.BaseRequest = "ac.UCloudBaseRequest"
	}

	for i := range action.Request {
		parseParameter(&action.Request[i])
	}

	for i := range action.Response {
		parseParameter(&action.Response[i])
	}

	t := template.Must(template.New(action.Name).Parse(actionControllerTpl))

	buff := bytes.NewBuffer(nil)
	if err := t.Execute(buff, action); err != nil {
		fmt.Printf("render template error %s\n", err)
		return err
	}
	code := formatCode(buff.Bytes())

	fp := path.Join(currpath, "controllers")
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		if err := os.MkdirAll(fp, 0777); err != nil {
			fmt.Printf("Could not create controllers directory: %s\n", err)
			return err
		}
	}
	fPath := path.Join(fp, action.Name+".go")
	f, err := os.OpenFile(fPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	f.Write(code)
	return nil
}

var (
	product string
	token   string
)

func GetGenerateCmd() *cobra.Command {
	generateCmd := &cobra.Command{
		Use:   "codegen",
		Short: "generate action api",
		Long:  ` generate action api for UCloud gateway, example apibae api <Action> --token <Cookie>`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				actions []*ActionDescribe
			)

			// 查询所有action

			req := httplib.NewBeegoRequest("https://uxiao.ucloudadmin.com/uqa/apisource", "GET")
			req.Param("product", product)
			req.SetCookie(&http.Cookie{Name: "INNER_AUTH_TOKEN", Value: token})
			var v struct {
				List []struct {
					Name string `json:"name"`
				} `json:"list"`
			}
			if err := req.ToJSON(&v); err != nil {
				fmt.Printf("get actions from uxiao return %s", err)
				return
			}
			for _, v := range v.List {
				req := httplib.NewBeegoRequest("https://uxiao.ucloudadmin.com/uqa/apisource", "GET")
				req.Param("api", v.Name)
				req.SetCookie(&http.Cookie{Name: "INNER_AUTH_TOKEN", Value: token})
				var v struct {
					List []ActionDescribe `json:"list"`
				}
				if err := req.ToJSON(&v); err != nil {
					fmt.Printf("get action from uxiao return %s", err)
					return
				}
				actions = append(actions, &v.List[0])
			}
			// 查询models
			req = httplib.NewBeegoRequest("https://uxiao.ucloudadmin.com/api_request/ListModel", "POST")
			req.JSONBody(map[string]interface{}{
				"Product": product,
			})
			req.SetCookie(&http.Cookie{Name: "INNER_AUTH_TOKEN", Value: token})
			var mv struct {
				Models []*Model
			}
			if err := req.ToJSON(&mv); err != nil {
				fmt.Printf("get actions from uxiao return %s", err)
				return
			}

			if err := GenerateModels(".", mv.Models); err != nil {
				fmt.Printf("GenerateAPI return %s", err)
			}

			for _, v := range actions {
				if err := GenerateActionController(".", v); err != nil {
					fmt.Printf("GenerateAPI return %s", err)
					return
				}
			}

		},
	}

	generateCmd.Flags().StringVarP(&product, "product", "p", "", "product name")
	generateCmd.Flags().StringVarP(&token, "token", "t", "", "get Action from uxiao.ucloudadmin.com with this cookies")
	return generateCmd
}
