package reddit

type response struct {
	Data data `json:"data"`
}

type data struct {
	Children []post `json:"children"`
}

type post struct {
	Data postdata `json:"data"`
}

type postdata struct {
	Title string `json:"title"`
	Ups   int    `json:"ups"`
	URL   string `json:"url"`
}
