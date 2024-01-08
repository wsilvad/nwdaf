package datacollection

import (
	"bytes"
	"fmt"
	"github.com/wsilvad/nwdaf/consumer"
	nwdaf_context "github.com/wsilvad/nwdaf/context"
	"github.com/wsilvad/nwdaf/util"
	"github.com/free5gc/openapi/Nnrf_NFDiscovery"
	"github.com/free5gc/openapi/models"
	"io/ioutil"
	"log"
	"net/http"
	
	"encoding/json"
	"github.com/wsilvad/nwdaf/model"

	"time"
	"strconv"

)


func InitEventExposureSubscriber(self*nwdaf_context.NWDAFContext) {

	searchOpt := Nnrf_NFDiscovery.SearchNFInstancesParamOpts{};
	// recupera todas as AMFs registradas na NRF
	resp, err := consumer.SendSearchNFInstances(self.NrfUri, models.NfType_AMF, models.NfType_NWDAF, searchOpt);
	if err != nil {
		fmt.Println(err)
	}

	//para cada uma das AMF's registrar no core realiza o subscriber de coleta
	for _, nfProfile := range resp.NfInstances {

		/* localiza a URL do end-point de subscriber com status de REGISTRADO */
		amfUri, endpoint, apiversion := util.SearchNFServiceUri(nfProfile, models.ServiceName_NAMF_EVTS, models.NfServiceStatus_REGISTERED)

		fmt.Println(endpoint)
		fmt.Println(apiversion)

		var buffer bytes.Buffer;

		buffer.WriteString(amfUri);
		buffer.WriteString("/");
		buffer.WriteString(endpoint);
		buffer.WriteString("/");
		buffer.WriteString(apiversion);
		buffer.WriteString("/");
		buffer.WriteString("subscriptions");

		url := buffer.String()

		/*
		 * 1 º os possiveis tipos de eventos p/ AMF estão em AmfEventType
		 */

		jsonData := `
    {	
		"Subscription" : { 	"EventList"	: 
										[{ "Type" : "REGISTRATION_ACCEPT",
                                           "ImmediateFlag" : true}], 
							"EventNotifyUri": "http://127.0.0.1:29599/datacollection/amf-contexts/registration-accept",
							"AnyUE" : true,
							"NfId"  : "NWDAF"
                          },
		"SupportedFeatures"	: "xx"
	}
		
	`

		var jsonStr = []byte(jsonData)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		req.Header.Set("X-Custom-Header", "myvalue")
		req.Header.Set("Content-Type", "application/json")

		client := util.GetHttpConnection()

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body) // response body is []byte
		fmt.Println(string(body))
	}
}

func InitEventExposureSubscriberPrometheus(self*nwdaf_context.NWDAFContext) {
	ue := []string{"192.168.10.22"}
	i := 0

	// cat tree times ago
	// now := time.Now()
	// threeHoursAgo := now.Add(-3 * time.Hour)

	// now_timestamp := now.Unix()
	// threeHoursAgo_timestamp := threeHoursAgo.Unix()

	// fmt.Printf("Now: %+v", now_timestamp)
	// fmt.Printf("Tree times Ago: %+v", threeHoursAgo_timestamp)

	for i < 1 {
		for num_ue := 0; num_ue < len(ue); num_ue ++ {
			fmt.Println("###### Calling Promtheus API ######")
			client := &http.Client{}
			// req, err := http.NewRequest("GET", "http://"+ue[num_ue]+":9090/api/v1/query_range?query=node_network_transmit_bytes_total%7Bdevice%3D%22gretun1%22%7D&start="+threeHoursAgo_timestamp+"&end="+now_timestamp+"&step=14&_=1703425983189&", nil)
			req, err := http.NewRequest("GET", "http://"+ue[num_ue]+":9090/api/v1/query_range?query=node_network_transmit_bytes_total%7Bdevice%3D%22gretun1%22%7D&start=1703422533.664&end=1703426133.664&step=14&_=1703425983189&", nil)
			if err != nil {
			fmt.Print(err.Error())
			}
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
			fmt.Print(err.Error())
			}
			defer resp.Body.Close()
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
			fmt.Print(err.Error())
			}
			var responseObject model.PrometheusResponseMain

			json.Unmarshal(bodyBytes, &responseObject)
			fmt.Printf("API Response as struct from %+v\n\n", responseObject.Data.Result[0].Metric)
			fmt.Printf("Data Reponse: %+v\n\n", responseObject.Data.Result[0].Values)

			util.AddPrometheusData(&responseObject)
		}

		time.Sleep(100 * time.Second)
	}
}