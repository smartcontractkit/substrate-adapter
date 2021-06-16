package adapter

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type handler func(req Request) (interface{}, error)

func RunWebserver(
	handler handler,
	port string,
) {
	if port == "" {
		port = "8080"
	}
	srv := NewHTTPService(handler)
	err := srv.Router.Run(fmt.Sprintf(":%v", port))
	if err != nil {
		fmt.Println(err)
	}
}

type HttpService struct {
	Router  *gin.Engine
	Handler handler
}

func NewHTTPService(
	handler handler,
) *HttpService {
	srv := HttpService{
		Router:  gin.Default(),
		Handler: handler,
	}
	srv.createRouter()
	return &srv
}

func (srv *HttpService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv.Router.ServeHTTP(w, r)
}

func (srv *HttpService) createRouter() {
	r := gin.Default()
	r.POST("/", srv.Call)

	srv.Router = r
}

type JobReq struct {
	JobID string  `json:"id"`
	Data  Request `json:"data"`
}

func validateRequest(t *JobReq) error {
	validations := []int{
		len(t.JobID),
	}
	fmt.Println(t.JobID)
	fmt.Println(t.Data)

	for _, v := range validations {
		if v == 0 {
			return errors.New("missing required field(s)")
		}
	}

	return nil
}

type resp struct {
	JobRunID   string      `json:"jobRunID"`
	StatusCode int         `json:"status_code"`
	Status     string      `json:"status"`
	Data       interface{} `json:"data"`
	Error      interface{} `json:"error"`
}

func errorJob(c *gin.Context, statusCode int, jobId, error string) {
	c.JSON(statusCode, resp{
		JobRunID:   jobId,
		StatusCode: statusCode,
		Status:     "errored",
		Error:      error,
	})
}

func (srv *HttpService) Call(c *gin.Context) {
	var req JobReq

	if err := c.BindJSON(&req); err != nil {
		log.Println(err)
		errorJob(c, http.StatusBadRequest, req.JobID, "Invalid JSON payload")
		return
	}
	// println("request: ")

	// body, _ := ioutil.ReadAll(c.Request.Response.Body)
	// body, _ := ioutil.Read(c.Request)

	// println(c.Request.Header)
	// println(string(body))

	if err := validateRequest(&req); err != nil {
		log.Println(err)
		errorJob(c, http.StatusBadRequest, req.JobID, err.Error())
		return
	}

	res, err := srv.Handler(req.Data)
	fmt.Printf("%+v\n", req.Data)

	if err != nil {
		log.Println(err)
		errorJob(c, http.StatusInternalServerError, req.JobID, err.Error())
		return
	}

	c.JSON(http.StatusOK, resp{
		JobRunID:   req.JobID,
		StatusCode: http.StatusOK,
		Status:     "success",
		Data:       res,
	})
}
