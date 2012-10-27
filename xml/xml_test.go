package xml

import (
	coreXML "encoding/xml"
	"io/ioutil"
	. "launchpad.net/gocheck"
	"reflect"
	"testing"
	"time"
)

type ECSResponse struct {
	//XMLName xml.Name `xml:"ItemLookupResponse"`
	Items []Item `xml:"Items>Item"`
}

type Item struct {
	ASIN           string
	SalesRank      int
	SmallImage     string `xml:"SmallImage>URL"`
	MediumImage    string `xml:"MediumImage>URL"`
	LargeImage     string `xml:"LargeImage>URL"`
	ItemAttributes ItemAttributes
	TotalOffers    int     `xml:"Offers>TotalOffers"`
	Offers         []Offer `xml:"Offers>Offer"`
}

type ImageSize struct {
	Size  int
	Units string `xml:"Units,attr"`
}

type ItemAttributes struct {
	Actor           string
	Binding         string
	EAN             string
	Label           string
	Manufacturer    string
	MPN             string
	NumberOfDiscs   int
	PackageQuantity int
	ProductGroup    string
	UPC             string
	Price           int    `xml:"ListPrice>Amount"`
	CurrencyCode    string `xml:"ListPrice>CurrencyCode"`
	ReleaseDate     string
	Title           string
}

type Offer struct {
	Merchant     Merchant
	Condition    string `xml:"OfferAttributes>Condition"`
	SubCondition string `xml:"OfferAttributes>SubCondition"`
	Price        int    `xml:"OfferListing>Price>Amount"`
	CurrencyCode string `xml:"OfferListing>Price>CurrencyCode"`
}

type Merchant struct {
	MerchantId            string
	Name                  string
	AverageFeedbackRating float32
	TotalFeedback         int
}

// Hook up gocheck into the gotest runner
func Test(t *testing.T) { TestingT(t) }

type lXMLSuite struct {
	ecs_xml []byte
}

var _ = Suite(&lXMLSuite{})

func (s *lXMLSuite) SetUpSuite(c *C) {
	var err error
	s.ecs_xml, err = ioutil.ReadFile("testdata/ecs.xml")
	if err != nil {
		c.Fatalf("ERROR: %v\n", err)
	}
}

func (s *lXMLSuite) BenchmarkCoreXML(c *C) {
	for i := 0; i < c.N; i++ {
		v := ECSResponse{}
		err := coreXML.Unmarshal(s.ecs_xml, &v)
		if err != nil {
			c.Fatalf("ERROR: %v\n", err)
		}
	}
}

func (s *lXMLSuite) BenchmarkLXML(c *C) {
	for i := 0; i < c.N; i++ {
		v := ECSResponse{}
		err := Unmarshal(s.ecs_xml, &v)
		if err != nil {
			c.Fatalf("ERROR: %v\n", err)
		}
	}
}

func (s *lXMLSuite) BenchmarkFlatLXML(c *C) {
	for i := 0; i < c.N; i++ {
		v := ECSResponse{}
		err := GokoUnmarshal(s.ecs_xml, &v)
		if err != nil {
			c.Fatalf("ERROR: %v\n", err)
		}
	}
}

func (s *lXMLSuite) TestCoreXMLvLXMLEqual(c *C) {
	res := ECSResponse{}
	err := Unmarshal(s.ecs_xml, &res)
	if err != nil {
		c.Fatalf("ERROR: %v\n", err)
	}

	res2 := ECSResponse{}
	err = coreXML.Unmarshal(s.ecs_xml, &res2)
	if err != nil {
		c.Fatalf("ERROR: %v\n", err)
	}

	c.Check(res, DeepEquals, res2)
}

const pathTestString = `
<Result>
    <Before>1</Before>
    <Items>
        <Item1>
            <Value>A</Value>
        </Item1>
        <Item2>
            <Value>B</Value>
        </Item2>
        <Item1>
            <Value>C</Value>
            <Value>D</Value>
        </Item1>
        <_>
            <Value>E</Value>
        </_>
    </Items>
    <After>2</After>
</Result>
`

type PathTestItem struct {
	Value string
}

type PathTestA struct {
	Items         []PathTestItem `xml:">Item1"`
	Before, After string
}

type PathTestB struct {
	Other         []PathTestItem `xml:"Items>Item1"`
	Before, After string
}

type PathTestC struct {
	Values1       []string `xml:"Items>Item1>Value"`
	Values2       []string `xml:"Items>Item2>Value"`
	Before, After string
}

type PathTestSet struct {
	Item1 []PathTestItem
}

type PathTestD struct {
	Other         PathTestSet `xml:"Items"`
	Before, After string
}

type PathTestE struct {
	Underline     string `xml:"Items>_>Value"`
	Before, After string
}

var pathTests = []interface{}{
	&PathTestA{Items: []PathTestItem{{"A"}, {"D"}}, Before: "1", After: "2"},
	&PathTestB{Other: []PathTestItem{{"A"}, {"D"}}, Before: "1", After: "2"},
	&PathTestC{Values1: []string{"A", "C", "D"}, Values2: []string{"B"}, Before: "1", After: "2"},
	&PathTestD{Other: PathTestSet{Item1: []PathTestItem{{"A"}, {"D"}}}, Before: "1", After: "2"},
	&PathTestE{Underline: "E", Before: "1", After: "2"},
}

// From encoding/xml/read_test.go
func (s *lXMLSuite) TestUnmarshalPaths(c *C) {
	for _, pt := range pathTests {
		v := reflect.New(reflect.TypeOf(pt).Elem()).Interface()
		if err := Unmarshal([]byte(pathTestString), v); err != nil {
			c.Fatalf("Unmarshal: %s", err)
		}

		c.Check(v, DeepEquals, pt)
	}
}

const atomFeedString = `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom" xml:lang="en-us" updated="2009-10-04T01:35:58+00:00"><title>Code Review - My issues</title><link href="http://codereview.appspot.com/" rel="alternate"></link><link href="http://codereview.appspot.com/rss/mine/rsc" rel="self"></link><id>http://codereview.appspot.com/</id><author><name>rietveld&lt;&gt;</name></author><entry><title>rietveld: an attempt at pubsubhubbub
</title><link href="http://codereview.appspot.com/126085" rel="alternate"></link><updated>2009-10-04T01:35:58+00:00</updated><author><name>email-address-removed</name></author><id>urn:md5:134d9179c41f806be79b3a5f7877d19a</id><summary type="html">
  An attempt at adding pubsubhubbub support to Rietveld.
http://code.google.com/p/pubsubhubbub
http://code.google.com/p/rietveld/issues/detail?id=155

The server side of the protocol is trivial:
  1. add a &amp;lt;link rel=&amp;quot;hub&amp;quot; href=&amp;quot;hub-server&amp;quot;&amp;gt; tag to all
     feeds that will be pubsubhubbubbed.
  2. every time one of those feeds changes, tell the hub
     with a simple POST request.

I have tested this by adding debug prints to a local hub
server and checking that the server got the right publish
requests.

I can&amp;#39;t quite get the server to work, but I think the bug
is not in my code.  I think that the server expects to be
able to grab the feed and see the feed&amp;#39;s actual URL in
the link rel=&amp;quot;self&amp;quot;, but the default value for that drops
the :port from the URL, and I cannot for the life of me
figure out how to get the Atom generator deep inside
django not to do that, or even where it is doing that,
or even what code is running to generate the Atom feed.
(I thought I knew but I added some assert False statements
and it kept running!)

Ignoring that particular problem, I would appreciate
feedback on the right way to get the two values at
the top of feeds.py marked NOTE(rsc).


</summary></entry><entry><title>rietveld: correct tab handling
</title><link href="http://codereview.appspot.com/124106" rel="alternate"></link><updated>2009-10-03T23:02:17+00:00</updated><author><name>email-address-removed</name></author><id>urn:md5:0a2a4f19bb815101f0ba2904aed7c35a</id><summary type="html">
  This fixes the buggy tab rendering that can be seen at
http://codereview.appspot.com/116075/diff/1/2

The fundamental problem was that the tab code was
not being told what column the text began in, so it
didn&amp;#39;t know where to put the tab stops.  Another problem
was that some of the code assumed that string byte
offsets were the same as column offsets, which is only
true if there are no tabs.

In the process of fixing this, I cleaned up the arguments
to Fold and ExpandTabs and renamed them Break and
_ExpandTabs so that I could be sure that I found all the
call sites.  I also wanted to verify that ExpandTabs was
not being used from outside intra_region_diff.py.


</summary></entry></feed> 	   `

type Feed struct {
	Title   string    `xml:"title"`
	Id      string    `xml:"id"`
	Link    []Link    `xml:"link"`
	Updated time.Time `xml:"updated,attr"`
	Author  Person    `xml:"author"`
	Entry   []Entry   `xml:"entry"`
}

type Entry struct {
	Title   string    `xml:"title"`
	Id      string    `xml:"id"`
	Link    []Link    `xml:"link"`
	Updated time.Time `xml:"updated"`
	Author  Person    `xml:"author"`
	Summary Text      `xml:"summary"`
}

type Link struct {
	Rel  string `xml:"rel,attr,omitempty"`
	Href string `xml:"href,attr"`
}

type Person struct {
	Name     string `xml:"name"`
	URI      string `xml:"uri"`
	Email    string `xml:"email"`
	InnerXML string `xml:",innerxml"`
}

type Text struct {
	Type string `xml:"type,attr,omitempty"`
	Body string `xml:",chardata"`
}

func ParseTime(str string) time.Time {
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		panic(err)
	}
	return t
}

var atomFeed = Feed{
	// TODO: Fix namespace support.
	// XMLName: Name{"http://www.w3.org/2005/Atom", "feed"},
	Title: "Code Review - My issues",
	Link: []Link{
		{Rel: "alternate", Href: "http://codereview.appspot.com/"},
		{Rel: "self", Href: "http://codereview.appspot.com/rss/mine/rsc"},
	},
	Id:      "http://codereview.appspot.com/",
	Updated: ParseTime("2009-10-04T01:35:58+00:00"),
	Author: Person{
		Name:     "rietveld<>",
		InnerXML: "<name>rietveld&lt;&gt;</name>",
	},
	Entry: []Entry{
		{
			Title: "rietveld: an attempt at pubsubhubbub\n",
			Link: []Link{
				{Rel: "alternate", Href: "http://codereview.appspot.com/126085"},
			},
			Updated: ParseTime("2009-10-04T01:35:58+00:00"),
			Author: Person{
				Name:     "email-address-removed",
				InnerXML: "<name>email-address-removed</name>",
			},
			Id: "urn:md5:134d9179c41f806be79b3a5f7877d19a",
			Summary: Text{
				Type: "html",
				Body: `
  An attempt at adding pubsubhubbub support to Rietveld.
http://code.google.com/p/pubsubhubbub
http://code.google.com/p/rietveld/issues/detail?id=155

The server side of the protocol is trivial:
  1. add a &lt;link rel=&quot;hub&quot; href=&quot;hub-server&quot;&gt; tag to all
     feeds that will be pubsubhubbubbed.
  2. every time one of those feeds changes, tell the hub
     with a simple POST request.

I have tested this by adding debug prints to a local hub
server and checking that the server got the right publish
requests.

I can&#39;t quite get the server to work, but I think the bug
is not in my code.  I think that the server expects to be
able to grab the feed and see the feed&#39;s actual URL in
the link rel=&quot;self&quot;, but the default value for that drops
the :port from the URL, and I cannot for the life of me
figure out how to get the Atom generator deep inside
django not to do that, or even where it is doing that,
or even what code is running to generate the Atom feed.
(I thought I knew but I added some assert False statements
and it kept running!)

Ignoring that particular problem, I would appreciate
feedback on the right way to get the two values at
the top of feeds.py marked NOTE(rsc).


`,
			},
		},
		{
			Title: "rietveld: correct tab handling\n",
			Link: []Link{
				{Rel: "alternate", Href: "http://codereview.appspot.com/124106"},
			},
			Updated: ParseTime("2009-10-03T23:02:17+00:00"),
			Author: Person{
				Name:     "email-address-removed",
				InnerXML: "<name>email-address-removed</name>",
			},
			Id: "urn:md5:0a2a4f19bb815101f0ba2904aed7c35a",
			Summary: Text{
				Type: "html",
				Body: `
  This fixes the buggy tab rendering that can be seen at
http://codereview.appspot.com/116075/diff/1/2

The fundamental problem was that the tab code was
not being told what column the text began in, so it
didn&#39;t know where to put the tab stops.  Another problem
was that some of the code assumed that string byte
offsets were the same as column offsets, which is only
true if there are no tabs.

In the process of fixing this, I cleaned up the arguments
to Fold and ExpandTabs and renamed them Break and
_ExpandTabs so that I could be sure that I found all the
call sites.  I also wanted to verify that ExpandTabs was
not being used from outside intra_region_diff.py.


`,
			},
		},
	},
}

func (s *lXMLSuite) TestCoreXMLvLXMLFeed(c *C) {
	var f1, f2 Feed
	if err := Unmarshal([]byte(atomFeedString), &f1); err != nil {
		c.Fatalf("Unmarshal: %s", err)
	}

	if err := coreXML.Unmarshal([]byte(atomFeedString), &f2); err != nil {
		c.Fatalf("Unmarshal: %s", err)
	}

	c.Check(f1, DeepEquals, f2)
}

// Stripped down Atom feed data structures.
// From encoding/xml/read_test.go
func (s *lXMLSuite) TestUnmarshalFeed(c *C) {
	var f Feed
	if err := Unmarshal([]byte(atomFeedString), &f); err != nil {
		c.Fatalf("Unmarshal: %s", err)
	}

	c.Check(f, DeepEquals, atomFeed)
}

const OK = "OK"
const withoutNameTypeData = `
<?xml version="1.0" charset="utf-8"?>
<Test3 Attr="OK" />`

type TestThree struct {
	XMLName coreXML.Name `xml:"Test3"`
	Attr    string       `xml:",attr"`
}

func (s *lXMLSuite) TestUnmarshalWithoutNameType(c *C) {
	var x TestThree
	if err := Unmarshal([]byte(withoutNameTypeData), &x); err != nil {
		c.Fatalf("Unmarshal: %s", err)
	}
	if x.Attr != OK {
		c.Fatalf("have %v\nwant %v", x.Attr, OK)
	}
}
