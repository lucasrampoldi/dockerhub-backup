package _package

//login
type Payload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token"`
}


//tags
type Image struct {
	Architecture string      `json:"architecture"`
	Features     interface{} `json:"features"`
	Variant      interface{} `json:"variant"`
	Digest       string      `json:"digest"`
	OS           string      `json:"os"`
	OSFeatures   interface{} `json:"os_features"`
	OSVersion    interface{} `json:"os_version"`
	Size         int64       `json:"size"`
}
type ResultTags struct {
	Creator             int64       `json:"creator"`
	ID                  int64       `json:"id"`
	ImageID             interface{} `json:"image_id"`
	Images              []Image     `json:"images"`
	LastUpdated         string      `json:"last_updated"`
	LastUpdater         int64       `json:"last_updater"`
	LastUpdaterUsername string      `json:"last_updater_username"`
	Name                string      `json:"name"`
	Repository          int64       `json:"repository"`
	FullSize            int64       `json:"full_size"`
	V2                  bool        `json:"v2"`
}
type Tags struct {
	Count    int64       `json:"count"`
	Next     interface{} `json:"next"`
	Previous interface{} `json:"previous"`
	Results  []ResultTags   `json:"results"`
}