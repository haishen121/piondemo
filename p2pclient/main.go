package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/pion/webrtc/v3"
)

func main() {
	done := make(chan struct{}, 1)
	var Candidates []string
	var IsOffer bool

	host := flag.String("h", "121.121.121.121:8080", "host:port")
	oi := flag.String("oi", "1", "oi")
	ai := flag.String("si", "2", "ai")
	flag.BoolVar(&IsOffer, "f", true, "false: answer  true:offer")

	flag.Parse()

	peer, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	peer.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			fmt.Println("candidate>null")
			b := bytes.Buffer{}
			for _, c := range Candidates {
				b.WriteString(c)
				b.WriteByte('\n')
			}

			user := *ai
			if IsOffer {
				user = *oi
			}
			resp, err := http.Post(fmt.Sprintf("http://%s/%s/candidates", *host, user), "application/txt; charset=utf-8", bytes.NewReader(b.Bytes()))
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			done <- struct{}{}
			return
		}
		Candidates = append(Candidates, c.ToJSON().Candidate)
		fmt.Printf("> %s\n", c.ToJSON().Candidate)
	})

	peer.OnICEConnectionStateChange(func(is webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State change %s\n", is.String())
		switch is {
		case webrtc.ICEConnectionStateClosed, webrtc.ICEConnectionStateDisconnected, webrtc.ICEConnectionStateFailed:
			os.Exit(0)
		}
	})

	if !IsOffer { //answer
		peer.OnDataChannel(func(dc *webrtc.DataChannel) {
			dc.OnOpen(func() {
				fmt.Printf("# Data Chan '%s/%d' open\n", dc.Label(), dc.ID())
				//send data
				for range time.NewTicker(5 * time.Second).C {
					now := time.Now()
					fmt.Printf("# C Data Chan Sen %s: %s\n", dc.Label(), now.String())
					if err := dc.Send([]byte(now.String())); err != nil {
						panic(err)
					}
				}
			})

			dc.OnMessage(func(msg webrtc.DataChannelMessage) {
				fmt.Printf("# C Data Chan Get %s: %s\n", dc.Label(), string(msg.Data))
			})
		})

	LoopForOfferSdp:
		for {
			resp, err := http.Get(fmt.Sprintf("http://%s/%s/sdp", *host, *oi))
			if err != nil || resp.StatusCode != 200 {
				time.Sleep(time.Second * 1)
				if err == nil {
					resp.Body.Close()
				}
				continue
			}
			remoteSdp := webrtc.SessionDescription{}

			if err := json.NewDecoder(resp.Body).Decode(&remoteSdp); err != nil {
				panic(err)
			}

			if err := peer.SetRemoteDescription(remoteSdp); err != nil {
				panic(err)
			}
			fmt.Printf("Get remoteSdp:%s\n%s\n\n", remoteSdp.Type, remoteSdp.SDP)

			resp.Body.Close()
			break LoopForOfferSdp
		}

		answer, err := peer.CreateAnswer(nil)
		if err != nil {
			panic(err)
		}
		if err = peer.SetLocalDescription(answer); err != nil {
			panic(err)
		}

		o, err := json.Marshal(answer)
		if err != nil {
			panic(err)
		}
		localSdp := string(o)
		fmt.Printf("answer:\n%s\n\n", localSdp)

		resp, err := http.Post(fmt.Sprintf("http://%s/%s/sdp", *host, *ai), "application/json; charset=utf-8", bytes.NewReader(o))
		if err != nil {
			panic(err)
		}
		if err = resp.Body.Close(); err != nil {
			panic(err)
		}

		<-done
	LoopForOfferCandidate:
		for {
			resp, err := http.Get(fmt.Sprintf("http://%s/%s/candidates", *host, *oi))
			if err != nil || resp.StatusCode != 200 {
				time.Sleep(time.Second * 1)
				if err == nil {
					resp.Body.Close()
				}
				continue
			}
			txt, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			for _, c := range bytes.Split(txt, []byte("\n")) {
				if err = peer.AddICECandidate(webrtc.ICECandidateInit{Candidate: string(c)}); err != nil {
					panic(err)
				}
			}

			resp.Body.Close()
			break LoopForOfferCandidate
		}

	} else { //offer
		dc, err := peer.CreateDataChannel("server", nil)
		if err != nil {
			panic(err)
		}
		dc.OnOpen(func() {
			fmt.Printf("# S Data Chan '%s/%d' open\n", dc.Label(), dc.ID())
			//send data
			for range time.NewTicker(5 * time.Second).C {
				now := time.Now()
				fmt.Printf("# S Data Chan Sen %s: %s\n", dc.Label(), now.String())
				if err := dc.Send([]byte(now.String())); err != nil {
					panic(err)
				}
			}
		})
		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("# S Data Chan Get %s: %s\n", dc.Label(), string(msg.Data))
		})

		offer, err := peer.CreateOffer(nil)
		if err != nil {
			panic(err)
		}
		if err = peer.SetLocalDescription(offer); err != nil {
			panic(err)
		}

		o, err := json.Marshal(offer)
		if err != nil {
			panic(err)
		}
		localSdp := string(o)

		fmt.Printf("offer:\n%s\n\n", localSdp)

		resp, err := http.Post(fmt.Sprintf("http://%s/%s/sdp", *host, *oi), "application/json; charset=utf-8", bytes.NewReader(o))
		if err != nil {
			panic(err)
		}
		if err = resp.Body.Close(); err != nil {
			panic(err)
		}

	LoopForAnswerSdp:
		for {
			resp, err := http.Get(fmt.Sprintf("http://%s/%s/sdp", *host, *ai))
			if err != nil || resp.StatusCode != 200 {
				time.Sleep(time.Second * 1)
				if err == nil {
					resp.Body.Close()
				}
				continue
			}
			remoteSdp := webrtc.SessionDescription{}

			if err := json.NewDecoder(resp.Body).Decode(&remoteSdp); err != nil {
				panic(err)
			}

			if err := peer.SetRemoteDescription(remoteSdp); err != nil {
				panic(err)
			}
			fmt.Printf("Get remoteSdp:%s\n%s\n\n", remoteSdp.Type, remoteSdp.SDP)

			resp.Body.Close()
			break LoopForAnswerSdp
		}

		<-done
	LoopForAnswerCandidate:
		for {
			resp, err := http.Get(fmt.Sprintf("http://%s/%s/candidates", *host, *ai))
			if err != nil || resp.StatusCode != 200 {
				time.Sleep(time.Second * 1)
				if err == nil {
					resp.Body.Close()
				}
				continue
			}
			txt, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			for _, c := range bytes.Split(txt, []byte("\n")) {
				if err = peer.AddICECandidate(webrtc.ICECandidateInit{Candidate: string(c)}); err != nil {
					panic(err)
				}
			}

			resp.Body.Close()
			break LoopForAnswerCandidate
		}
	}

	select {}
}
