package vitotrol

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	td "github.com/maxatome/go-testdeep"
)

type testAction struct {
	expectedRequest interface{}
	serverResponse  string
}

func testSendRequestAnyMulti(t *td.T,
	sendReqs func(v *Session, d *Device) bool,
	actions map[string]*testAction, testName string) bool {
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			soapActionURL := r.Header.Get("SOAPAction")
			if !t.NotEmpty(soapActionURL,
				"%s: SOAPAction header found", testName) {
				w.WriteHeader(http.StatusNotAcceptable)
				return
			}
			soapAction := soapActionURL[strings.LastIndex(soapActionURL, "/")+1:]
			pAction := actions[soapAction]
			if !t.NotNil(pAction,
				"%s: SOAPAction header `%s' matches one expected action",
				testName, soapAction) {
				w.WriteHeader(http.StatusNotAcceptable)
				return
			}

			t.CmpDeeply(r.Header.Get("Content-Type"), "text/xml; charset=utf-8",
				"%s: Content-Type header matches", testName)

			if cookie := r.Header.Get("Cookie"); cookie != "" {
				w.Header().Add("Set-Cookie", cookie)
			}

			// Extract request body in the same struct type as the expectedRequest
			recvReq := virginInstance(pAction.expectedRequest)
			if !extractRequestBody(t, r, recvReq, testName) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			t.CmpDeeply(recvReq, pAction.expectedRequest, "%s: request OK", testName)

			// Send response
			fmt.Fprintln(w, respHeader+pAction.serverResponse+respFooter)
		}))
	defer ts.Close()

	MainURL = ts.URL
	v := &Session{
		Devices: []Device{
			{
				DeviceID:   testDeviceID,
				LocationID: testLocationID,
				Attributes: map[AttrID]*Value{},
				Timesheets: map[TimesheetID]map[string]TimeslotSlice{},
			},
		},
	}
	return sendReqs(v, &v.Devices[0])
}

//
// WriteDataWait
//

func TestWriteDataWait(tt *testing.T) {
	t := td.NewT(tt)

	// No problem
	testSendRequestAnyMulti(t,
		func(v *Session, d *Device) bool {
			WriteDataWaitDuration = 0
			WriteDataWaitMinDuration = 0
			ch, err := d.WriteDataWait(v, writeDataTestID, writeDataTestValue)
			if !t.CmpNoError(err) {
				return false
			}
			timeoutTicker := time.NewTicker(100 * time.Millisecond)
			defer timeoutTicker.Stop()

			select {
			case err = <-ch:
				return t.CmpNoError(err)
			case <-timeoutTicker.C:
				t.Error("TIMEOUT!")
				return false
			}
		},
		map[string]*testAction{
			"WriteData": {
				expectedRequest: writeDataTest.expectedRequest,
				serverResponse: intoDeviceResponse(
					"WriteData", writeDataTest.serverResponse),
			},
			"RequestWriteStatus": &requestWriteStatusTest,
		},
		"WriteDataWait")

	// Error during WriteData
	testSendRequestAnyMulti(t,
		func(v *Session, d *Device) bool {
			WriteDataWaitDuration = 0
			ch, err := d.WriteDataWait(v, writeDataTestID, writeDataTestValue)
			t.CmpError(err)
			return t.Nil(ch)
		},
		map[string]*testAction{
			"WriteData": {
				expectedRequest: writeDataTest.expectedRequest,
				serverResponse:  `<bad XML>`,
			},
			"RequestWriteStatus": &requestWriteStatusTest,
		},
		"WriteDataWait, error during WriteData")

	// Error during RequestWriteStatus
	testSendRequestAnyMulti(t,
		func(v *Session, d *Device) bool {
			ch, err := d.WriteDataWait(v, writeDataTestID, writeDataTestValue)
			if !t.CmpNoError(err) {
				return false
			}
			timeoutTicker := time.NewTicker(100 * time.Millisecond)
			defer timeoutTicker.Stop()

			select {
			case err = <-ch:
				return t.CmpError(err)
			case <-timeoutTicker.C:
				t.Error("TIMEOUT!")
				return false
			}
		},
		map[string]*testAction{
			"WriteData": {
				expectedRequest: writeDataTest.expectedRequest,
				serverResponse: intoDeviceResponse(
					"WriteData", writeDataTest.serverResponse),
			},
			"RequestWriteStatus": {
				expectedRequest: requestWriteStatusTest.expectedRequest,
				serverResponse:  `<bad XML>`,
			},
		},
		"WriteDataWait, error during RequestWriteStatus")
}

//
// RefreshDataWait
//

func TestRefreshDataWait(tt *testing.T) {
	t := td.NewT(tt)

	// No problem
	testSendRequestAnyMulti(t,
		func(v *Session, d *Device) bool {
			RefreshDataWaitDuration = 0
			RefreshDataWaitMinDuration = 0
			ch, err := d.RefreshDataWait(v, refreshDataTestIDs)
			if !t.CmpNoError(err) {
				return false
			}
			timeoutTicker := time.NewTicker(100 * time.Millisecond)
			defer timeoutTicker.Stop()

			select {
			case err = <-ch:
				return t.CmpNoError(err)
			case <-timeoutTicker.C:
				t.Error("TIMEOUT!")
				return false
			}
		},
		map[string]*testAction{
			"RefreshData": {
				expectedRequest: refreshDataTest.expectedRequest,
				serverResponse: intoDeviceResponse(
					"RefreshData", refreshDataTest.serverResponse),
			},
			"RequestRefreshStatus": &requestRefreshStatusTest,
		},
		"RefreshDataWait")

	// Error during RefreshData
	testSendRequestAnyMulti(t,
		func(v *Session, d *Device) bool {
			RefreshDataWaitDuration = 0
			ch, err := d.RefreshDataWait(v, refreshDataTestIDs)
			t.CmpError(err)
			return t.Nil(ch)
		},
		map[string]*testAction{
			"RefreshData": {
				expectedRequest: refreshDataTest.expectedRequest,
				serverResponse:  `<bad XML>`,
			},
			"RequestRefreshStatus": &requestRefreshStatusTest,
		},
		"RefreshDataWait, error during RefreshData")

	// Error during RequestRefreshStatus
	testSendRequestAnyMulti(t,
		func(v *Session, d *Device) bool {
			ch, err := d.RefreshDataWait(v, refreshDataTestIDs)
			if !t.CmpNoError(err) {
				return false
			}
			timeoutTicker := time.NewTicker(100 * time.Millisecond)
			defer timeoutTicker.Stop()

			select {
			case err = <-ch:
				return t.CmpError(err)
			case <-timeoutTicker.C:
				t.Error("TIMEOUT!")
				return false
			}
		},
		map[string]*testAction{
			"RefreshData": {
				expectedRequest: refreshDataTest.expectedRequest,
				serverResponse: intoDeviceResponse(
					"RefreshData", refreshDataTest.serverResponse),
			},
			"RequestRefreshStatus": {
				expectedRequest: requestRefreshStatusTest.expectedRequest,
				serverResponse:  `<bad XML>`,
			},
		},
		"RefreshDataWait, error during RequestRefreshStatus")
}
