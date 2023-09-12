package handlers

import (
	"github.com/stretchr/testify/assert"

	"net/http"
	"net/http/httptest"
	"testing"
)


func TestUpdateHandler1(t *testing.T) {
	type want struct {
        code        int     
        contentType string		
    }

	tests := []struct {
        name string
		url string
        want want
    }{
        {
            name: "positive test #1",
			url: "/update/counter/Alloc/22",
            want: want{
                code:        http.StatusOK,                
                contentType: "text/plain; charset=utf-8",				
            },
        },
		{
            name: "positive test #2",
			url: "/update/gauge/Alloc/11",
            want: want{
                code:        http.StatusOK,                
                contentType: "text/plain; charset=utf-8",				
            },
        },
        {
            name: "negative test #3",
			url: "/update/gauge/11",
            want: want{
                code:        http.StatusBadRequest,                
                contentType: "text/plain; charset=utf-8",				
            },
        },
        {
            name: "negative test #4",
			url: "/update/gauge/Alloc/value",
            want: want{
                code:        http.StatusBadRequest,                
                contentType: "text/plain; charset=utf-8",				
            },
        },
        {
            name: "negative test #5",
			url: "/",
            want: want{
                code:        http.StatusNotFound,                
                contentType: "text/plain; charset=utf-8",				
            },
        },
    }

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.url, nil)
			w := httptest.NewRecorder()
            UpdateHandler(w, request)
			
			res := w.Result()
			assert.Equal(t, res.Header.Get("Content-Type"), test.want.contentType)
			assert.Equal(t, res.StatusCode, test.want.code)		
			
		})
	}
}
