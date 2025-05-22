package cevio_server

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/gotti/cevigo/pkg/cevioai"
	"github.com/spf13/pflag"
)

type talkcast struct {
	Name         string      `json:"name"`
	Speaker_uuid string      `json:"speaker_uuid"`
	Style        []talkstyle `json:"styles"`
}

type talkstyle struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type talkstyletable struct {
	id    int
	name  string
	intid int
	key   string
}

type cevioServer struct {
	talker cevioai.ITalker2V40
	mtx    sync.Mutex
}
type cevioParams struct {
	Id              int     `json:"id"`
	Text            string  `json:"text"`
	SpeedScale      float32 `json:"speedScale"`
	PitchScale      float32 `json:"pitchScale"`
	IntonationScale float32 `json:"intonationScale"`
	VolumeScale     float32 `json:"volumeScale"`
}

var (
	table     []talkstyletable
	allcasts  []talkcast
	debugmode bool
	port      string
	apiname   string
)

func (s *cevioServer) allspeakers() {
	casts, _ := s.talker.GetAvailableCasts()
	array, _ := casts.ToGoArray()
	allcasts = []talkcast{}
	table = []talkstyletable{}
	i := 0
	for _, castname := range array {
		s.talker.SetCast(castname)
		cevistyle := []talkstyle{}
		compomemts, _ := s.talker.GetComponents()
		stylelength, _ := compomemts.GetLength()
		for styleindex := 0; styleindex < stylelength; styleindex++ {
			stylestruct, _ := compomemts.GetAt(styleindex)
			stylename, _ := stylestruct.GetName()
			table = append(table, talkstyletable{i, castname, styleindex, stylename})
			cevistyle = append(cevistyle, talkstyle{i, stylename})
			i += 1
		}
		allcasts = append(allcasts, talkcast{castname, "none", cevistyle})
	}
	if debugmode {
		fmt.Printf("%#v\n", allcasts)
	}
}

func (s *cevioServer) speakers(w http.ResponseWriter, r *http.Request) {
	js, _ := json.Marshal(allcasts)
	fmt.Fprint(w, string(js))
}

func (s *cevioServer) synthesis(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var param cevioParams
	json.NewDecoder(r.Body).Decode(&param)
	r.ParseForm()
	id, _ := strconv.ParseInt(r.Form.Get("speaker"), 10, 32)
	param.Id = int(id)
	param.PitchScale += 1
	param.SpeedScale *= 50
	param.PitchScale *= 50
	param.IntonationScale *= 50
	param.VolumeScale *= 50
	audio, _ := s.CreateWav(&param)
	w.Header().Set("Content-Disposition", "attachment; filename=query.wav")
	w.Header().Set("Content-Type", "audio/wav")
	w.WriteHeader(http.StatusOK)
	w.Write(audio)
	if debugmode {
		fmt.Printf("%v\n", param)
	}
}
func (s *cevioServer) audio_query(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Fprint(w, `{"id":0,"text":"`+r.Form.Get("text")+`","speedScale":1.0,"pitchScale":0.0,"intonationScale":1.0,"volumeScale":1.0}`)
}

func (s *cevioServer) applyParameters(p *cevioParams) error {
	s.talker.SetCast(table[int(p.Id)].name)
	s.talker.SetVolume(int(p.VolumeScale))
	s.talker.SetSpeed(int(p.SpeedScale))
	s.talker.SetTone(int(p.PitchScale))           //高さ
	s.talker.SetAlpha(int(50))                    //声質
	s.talker.SetToneScale(int(p.IntonationScale)) //抑揚
	compomemts, err := s.talker.GetComponents()
	stylelength, _ := compomemts.GetLength()
	if err != nil {
		return fmt.Errorf("getting components: %w", err)
	}
	for styleindex := 0; styleindex < stylelength; styleindex++ {
		n, _ := compomemts.GetAt(styleindex)
		n.SetValue(int(0))
		if styleindex == table[int(p.Id)].intid {
			n.SetValue(int(100))
		}
	}
	return nil
}

func (s *cevioServer) speak(text string) ([]byte, error) {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("generating rand: %w", err)
	}
	fPath := fmt.Sprintf("%8x", buf)
	fPath = filepath.Join(filepath.Join(os.TempDir(), "cevigo"), fPath)
	err = os.MkdirAll(filepath.Dir(fPath), os.FileMode(0755))
	if err != nil {
		return nil, fmt.Errorf("making dir: %w", err)
	}
	b, err := s.talker.OutputWaveToFile(text, fPath)
	if err != nil {
		return nil, err
	}
	if !b {
		return nil, fmt.Errorf("failed to outputting wav, please check the packet you sent\n error string: %v", text)
	}
	defer os.Remove(fPath)
	if err != nil {
		return nil, fmt.Errorf("outputting: %w", err)
	}
	f, err := os.Open(fPath)
	audio, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return audio, nil
}

func (s *cevioServer) CreateWav(req *cevioParams) ([]byte, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if err := s.applyParameters(req); err != nil {
		log.Printf("applying parameters: %v", err)
		return nil, err
	}
	audio, err := s.speak(req.Text)
	if err != nil {
		log.Printf("speaking: %v", err)
		return nil, err
	}
	return audio, nil
}

func Mainhttp() {
	pflag.StringVarP(&port, "port", "p", "10001", "10001,or 10002")
	pflag.StringVarP(&apiname, "api", "a", "cevio", "cevio, or cevioai")
	pflag.BoolVarP(&debugmode, "debug", "d", false, "True, or False")
	pflag.Parse()

	var apiname string
	if apiname == "cevio" {
		apiname = cevioai.CevioApiName
	} else if apiname == "cevioai" {
		apiname = cevioai.CevioAiApiName
	} else {
		log.Println("set cevio or cevioai to --api")
		os.Exit(1)
	}

	talker := cevioai.NewITalker2V40(apiname)
	log.Printf("connected to %s\n", apiname)
	s := &cevioServer{talker: talker}
	s.allspeakers()
	http.HandleFunc("/speakers", s.speakers)       // speakersを出力
	http.HandleFunc("/audio_query", s.audio_query) // オーディオクエリを出力
	http.HandleFunc("/synthesis", s.synthesis)     // 音声wavを出力
	http.ListenAndServe(":"+port, nil)
}
