package handler

import (
	"fmt"
	"net/http"
	dblayer "objectstore-server/db"
	"objectstore-server/util"
	"os"
	"time"
)

const (
	pwd_salt   = "#ws0"
	valid_time = 1000
)

type Data struct {
	Token    string `json:"Token,omitempty"`
	UserName string `json:"Username,omitempty"`
	Location string `json:"Location"`
}

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

	enc_passwd := util.Sha1([]byte(password + pwd_salt))
	suc := dblayer.UserSignUp(username, enc_passwd)
	if suc {
		w.Write([]byte("SUCCESS"))
		return
	}
	w.Write([]byte("FAILED"))
}

// SigninHandler 用户登录接口
func SigninHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	encPwd := util.Sha1([]byte(password + pwd_salt))

	// 校验用户名与密码
	check := dblayer.UserLogin(username, encPwd)
	if !check {
		w.Write([]byte("FAILED"))
		return
	}

	// 生成token
	token := getToken(username)
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		w.Write([]byte("FAILED"))
		return
	}

	// 登录成功后重定向
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: Data{
			Token:    token,
			UserName: username,
			Location: "/static/view/home.html",
		},
	}
	w.Write(resp.JSONBytes())
}

func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// 解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	token := r.Form.Get("token")

	// 验证token是否有效
	isValid := isTokenValid(token)
	if !isValid {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	// 查询用户信息
	info, err := dblayer.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: info,
	}
	w.Write(resp.JSONBytes())
}

func getToken(username string) string {
	// 40bit md5(username + timestamp + token_salt) + timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.Md5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

func isTokenValid(token string) bool {
	if len(token) != 40 {
		return false
	}
	//TODO: 判断token的时效性，是否过期
	//timeStamp, err := strconv.Atoi(token[32:])
	//if err != nil {
	//	return false
	//}
	//now := time.Now().Unix()
	//if now - timeStamp >
	//TODO: 从数据库表tbl_user_token查询username对应的token信息
	//TODO: 对比两个token是否一致
	return true
}
