package xml

import (
	gokoxml "github.com/moovweb/gokogiri/xml"
	"strconv"
)

func GokoUnmarshal(data []byte, result *ECSResponse) error {
	return new(GokoParser).Decode(data, result)
}

type GokoParser struct {
	doc *gokoxml.XmlDocument
}

func (p *GokoParser) Decode(data []byte, result *ECSResponse) error {
	doc, err := gokoxml.Parse(data, gokoxml.DefaultEncodingBytes, nil, gokoxml.DefaultParseOption, gokoxml.DefaultEncodingBytes)
	p.doc = doc
	if err != nil {
		return err
	}
	defer p.doc.Free()

	doc.XPathCtx.RegisterNamespace("ecs", "http://webservices.amazon.com/AWSECommerceService/2010-11-01")

	root := doc.Root()

	res, err := root.Search("ecs:Items/ecs:Item")
	if err == nil && len(res) >= 1 {
		for _, v := range res {
			result.Items = append(result.Items, p.parseItem(v.FirstChild()))
		}
	}

	return err
}

func (p *GokoParser) parseItem(a_node gokoxml.Node) Item {
	i := Item{}
	for cur_node := a_node; cur_node != nil; cur_node = cur_node.NextSibling() {
		if cur_node.NodeType() != gokoxml.XML_ELEMENT_NODE {
			continue
		}

		switch cur_node.Name() {
		case "ASIN":
			i.ASIN = cur_node.Content()
		case "SalesRank":
			i.SalesRank, _ = strconv.Atoi(cur_node.Content())
		// case "SmallImage":
		// 	i.SmallImage = cur_node.Content()
		case "MediumImage":
			i.MediumImage = cur_node.Content()
		case "LargeImage":
			i.LargeImage = cur_node.Content()
		case "ItemAttributes":
			i.ItemAttributes = p.parseItemAttribute(cur_node.FirstChild())
		case "Offers":
			res, err := cur_node.Search("ecs:TotalOffers")
			if err == nil && len(res) == 1 {
				i.TotalOffers, _ = strconv.Atoi(res[0].Content())
			}

			i.Offers = p.parseOffers(cur_node.FirstChild())
		}
	}

	return i
}

func (p *GokoParser) parseItemAttribute(a_node gokoxml.Node) ItemAttributes {
	i := ItemAttributes{}
	for cur_node := a_node; cur_node != nil; cur_node = cur_node.NextSibling() {
		if cur_node.NodeType() != gokoxml.XML_ELEMENT_NODE {
			continue
		}

		switch cur_node.Name() {
		case "Actor":
			i.Actor = cur_node.Content()
		case "Binding":
			i.Binding = cur_node.Content()
		case "EAN":
			i.EAN = cur_node.Content()
		case "Label":
			i.Label = cur_node.Content()
		case "Manufacturer":
			i.Manufacturer = cur_node.Content()
		case "MPN":
			i.MPN = cur_node.Content()
		case "ProductGroup":
			i.ProductGroup = cur_node.Content()
		case "UPC":
			i.UPC = cur_node.Content()
		case "CurrencyCode":
			i.CurrencyCode = cur_node.Content()
		case "ReleaseDate":
			i.ReleaseDate = cur_node.Content()
		case "Title":
			i.Title = cur_node.Content()
		case "NumberOfDiscs":
			i.NumberOfDiscs, _ = strconv.Atoi(cur_node.Content())
		case "PackageQuantity":
			i.PackageQuantity, _ = strconv.Atoi(cur_node.Content())
		case "ListPrice":
			res, err := cur_node.Search("ecs:Amount")
			if err == nil && len(res) == 1 {
				i.Price, _ = strconv.Atoi(res[0].Content())
			}
			res, err = cur_node.Search("ecs:CurrencyCode")
			if err == nil && len(res) == 1 {
				i.CurrencyCode = res[0].Content()
			}
		}
	}

	return i
}

func (p *GokoParser) parseOffers(a_node gokoxml.Node) []Offer {
	offers := []Offer{}

	for os_node := a_node; os_node != nil; os_node = os_node.NextSibling() {
		if os_node.NodeType() != gokoxml.XML_ELEMENT_NODE {
			continue
		}

		if os_node.Name() == "Offer" {

			i := Offer{}

			for cur_node := os_node.FirstChild(); cur_node != nil; cur_node = cur_node.NextSibling() {
				if cur_node.NodeType() != gokoxml.XML_ELEMENT_NODE {
					continue
				}

				switch cur_node.Name() {
				case "Merchant":
					for o_node := cur_node.FirstChild(); o_node != nil; o_node = o_node.NextSibling() {
						if o_node.NodeType() != gokoxml.XML_ELEMENT_NODE {
							continue
						}

						switch o_node.Name() {
						case "MerchantId":
							i.Merchant.MerchantId = o_node.Content()
						case "Name":
							i.Merchant.Name = o_node.Content()
						case "AverageFeedbackRating":
							v, _ := strconv.ParseFloat(o_node.Content(), 32)
							i.Merchant.AverageFeedbackRating = float32(v)
						case "TotalFeedback":
							i.Merchant.TotalFeedback, _ = strconv.Atoi(o_node.Content())
						}
					}

				case "OfferAttributes":
					for o_node := cur_node.FirstChild(); o_node != nil; o_node = o_node.NextSibling() {
						if o_node.NodeType() != gokoxml.XML_ELEMENT_NODE {
							continue
						}

						switch o_node.Name() {
						case "Condition":
							i.Condition = o_node.Content()
						case "SubCondition":
							i.SubCondition = o_node.Content()
						}
					}

				case "OfferListing":
					res, err := cur_node.Search("ecs:Price/ecs:Amount")
					if err == nil && len(res) == 1 {
						i.Price, _ = strconv.Atoi(res[0].Content())
					}
					res, err = cur_node.Search("ecs:Price/ecs:CurrencyCode")
					if err == nil && len(res) == 1 {
						i.CurrencyCode = res[0].Content()
					}
				}
			}

			offers = append(offers, i)
		}
	}

	return offers
}
