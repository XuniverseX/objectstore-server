package handler

import (
	"net/http"
	dblayer "objectstore-server/db"
	"objectstore-server/util"
	"os"
)

const (
	pwd_salt = "#wsh0"
)

// SignupHandler 处理用户注册接口
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// 跳转注册界面
		data, err := os.ReadFile("./static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
		return
	}

	r.ParseForm()

	username := r.Form.Get("username")
	password := r.Form.Get("password")

	enc_passwd := util.Md5([]byte(password + pwd_salt))
	suc := dblayer.UserSignUp(username, enc_passwd)
	if suc {
		w.Write([]byte("SUCCESS"))
		return
	}
	w.Write([]byte("FAILED"))
}
