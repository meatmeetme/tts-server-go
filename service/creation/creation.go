package creation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var (
	tokenUrl  = "https://southeastasia.customvoice.api.speech.microsoft.com/api/texttospeech/v3.0-beta1/accdemopageentry/auth-token"
	voicesUrl = "https://southeastasia.customvoice.api.speech.microsoft.com/api/texttospeech/v3.0-beta1/accdemopage/voices"
	speakUrl  = "https://southeastasia.customvoice.api.speech.microsoft.com/api/texttospeech/v3.0-beta1/accdemopage/speak"

	TokenErr = errors.New("unauthorized")
)

type Creation struct {
	token string
}

func (c *Creation) GetAudio(text, voiceName, rate, style, styleDegree, format string) ([]byte, error) {
	if c.token == "" {
		s, err := token()
		if err != nil {
			return nil, err
		}
		c.token = s
	}

	data, err := speak(c.token, text, voiceName, rate, style, styleDegree, format)
	if errors.Is(err, TokenErr) { /* Token已失效 */
		c.token = ""
		data, err = c.GetAudio(text, voiceName, rate, style, styleDegree, format)
	}

	return data, err
}

func speak(token, text, voiceName, rate, style, styleDegree, format string) ([]byte, error) {
	ssml := `<!--ID=B7267351-473F-409D-9765-754A8EBCDE05;Version=1|{\"VoiceNameToIdMapItems\":[{\"Id\":\"5f55541d-c844-4e04-a7f8-1723ffbea4a9\",\"Name\":\"Microsoft Server Speech Text to Speech Voice (zh-CN, XiaoxiaoNeural)\",\"ShortName\":\"` + voiceName + `\",\"Locale\":\"zh-CN\",\"VoiceType\":\"StandardVoice\"}]}-->\n<!--ID=5B95B1CC-2C7B-494F-B746-CF22A0E779B7;Version=1|{\"Locales\":{\"zh-CN\":{\"AutoApplyCustomLexiconFiles\":[{}]}}}-->\n<speak version=\"1.0\" xmlns=\"http://www.w3.org/2001/10/synthesis\" xmlns:mstts=\"http://www.w3.org/2001/mstts\" xmlns:emo=\"http://www.w3.org/2009/10/emotionml\" xml:lang=\"zh-CN\"><voice name=\"` + voiceName + `\"><mstts:express-as style=\"` + style + `\" styledegree=\"` + styleDegree + `\"><prosody rate=\"` + rate + `\" contour=\"\">` + text + `</prosody></mstts:express-as></voice></speak>`
	payload := strings.NewReader(`{
    "ssml": "` + ssml + `",
    "ttsAudioFormat": "` + format + `",
    "offsetInPlainText": 0,
    "lengthInPlainText":` + strconv.FormatInt(int64(len(text)), 10) +
		`,"properties": {
        "SpeakTriggerSource": "AccTuningPagePlayButton"
    }
}`)
	req, err := http.NewRequest(http.MethodPost, speakUrl, payload)

	if err != nil {
		return nil, err
	}
	req.Header.Add("AccDemoPageAuthToken", token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, TokenErr
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK { /* 服务器返回错误 大概率是SSML格式有问题 */
		return nil, errors.New(string(data))
	}

	return data, nil
}

func voices(token string) error {
	payload := strings.NewReader(`{"queryCondition":{"items":[{"name":"VoiceTypeList","value":"StandardVoice","operatorKind":"Contains"}]}}`)

	req, err := http.NewRequest(http.MethodPost, voicesUrl, payload)

	if err != nil {
		return err
	}
	req.Header.Add("AccDemoPageAuthToken", token)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36")
	req.Header.Add("X-Ms-Useragent", "SpeechStudio/2021.05.001")
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return nil
}

func token() (string, error) {
	resp, err := http.Get(tokenUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	value := make(map[string]string)
	err = json.NewDecoder(resp.Body).Decode(&value)
	if err != nil {
		return "", err
	}

	return value["authToken"], nil
}
