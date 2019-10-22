package commands

import (
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/patrickmn/go-cache"
)

var regionCodes = map[string]string{
	"ap-east-1":      "APE1",
	"ap-northeast-3": "APN3",
	"ap-northeast-1": "APN1",
	"ap-northeast-2": "APN2",
	"ap-southeast-1": "APS1",
	"ap-southeast-2": "APS2",
	"ap-south-1":     "APS3",
	"cn-north-1":     "CNN1",
	"cn-northwest-1": "CNW1",
	"ca-central-1":   "CAN1",
	"eu-north-1":     "EUN1",
	"eu-central-1":   "EUC1",
	"eu-west-1":      "EU",
	"eu-west-2":      "EUW2",
	"eu-west-3":      "EUW3",
	"sa-east-1":      "SAE1",
	"me-south-1":     "MES1",
	"us-gov-west-1":  "UGW1",
	"us-gov-east-1":  "UGE1",
	"us-east-1":      "USE1",
	"us-east-2":      "USE2",
	"us-west-1":      "USW1",
	"us-west-2":      "USW2",
}

type storageType struct {
	volumeType   string
	shortCode    string
	storageClass string
}

func init() {
	memCache = cache.New(5*time.Minute, 10*time.Minute)

}

var (
	memCache     *cache.Cache
	storageTypes = map[string]storageType{
		"STANDARD":            storageType{volumeType: "Standard", shortCode: "", storageClass: "General Purpose"},
		"REDUCED_REDUNDANCY":  storageType{volumeType: "Reduced Redundancy", shortCode: "RRS", storageClass: "Non-Critical Data"},
		"STANDARD_IA":         storageType{volumeType: "Standard - Infrequent Access", shortCode: "SIA", storageClass: "Infrequent Access"},
		"ONEZONE_IA":          storageType{volumeType: "One Zone - Infrequent Access", shortCode: "ZIA", storageClass: "Infrequent Access"},
		"INTELLIGENT_TIERING": storageType{volumeType: "Intelligent-Tiering Infrequent Access", shortCode: "INT-IA", storageClass: "Intelligent-Tiering"},
		"GLACIER":             storageType{volumeType: "Amazon Glacier", shortCode: "Glacier", storageClass: "Archive"},
		"DEEP_ARCHIVE":        storageType{volumeType: "Glacier Deep Archive", shortCode: "GDA", storageClass: "Staging"},
	}
)

func getByteHrsPrice(sess *session.Session, bucketRegion string, objStorageClass *string) (usd float64) {
	svc := pricing.New(sess, &aws.Config{
		Region: aws.String("ap-south-1"),
	})

	if objStorageClass == nil {
		standardClass := "STANDARD"
		objStorageClass = &standardClass
	}

	usagetype, storageClass, volumeType := getPricingFilters(bucketRegion, *objStorageClass)

	//Check cache
	price, found := memCache.Get(usagetype + storageClass + volumeType)
	if found {
		usd = price.(float64)
		return
	}

	usageFilter := &pricing.Filter{}
	usageFilter.SetField("usagetype")
	usageFilter.SetType("TERM_MATCH")
	usageFilter.SetValue(usagetype)

	storageClassFilter := &pricing.Filter{}
	storageClassFilter.SetField("storageClass")
	storageClassFilter.SetType("TERM_MATCH")
	storageClassFilter.SetValue(storageClass)

	volumeTypeFilter := &pricing.Filter{}
	volumeTypeFilter.SetField("volumeType")
	volumeTypeFilter.SetType("TERM_MATCH")
	volumeTypeFilter.SetValue(volumeType)

	input := &pricing.GetProductsInput{
		Filters: []*pricing.Filter{usageFilter, storageClassFilter, volumeTypeFilter},
	}
	input.SetServiceCode("AmazonS3")
	resp, err := svc.GetProducts(input)
	if err != nil {
		log.Fatalf("GetProducts API %s", err)
	}

	for _, list := range resp.PriceList {
		//product := list["product"].(map[string]interface{})
		//productSKU := product["sku"]
		terms := list["terms"].(map[string]interface{})["OnDemand"].(map[string]interface{})
		//fmt.Println("SKU", productSKU)
		for _, SKUOffer := range terms {
			offer := SKUOffer.(map[string]interface{})
			priceDimensions := offer["priceDimensions"].(map[string]interface{})
			for _, priceDimension := range priceDimensions {
				pricePerUnit := priceDimension.(map[string]interface{})["pricePerUnit"]
				usdPrice := pricePerUnit.(map[string]interface{})["USD"]
				usd, err := strconv.ParseFloat(usdPrice.(string), 64)
				if err != nil {
					log.Fatalf("GetProducts API %s", err)
				}
				//Cache Price for the filter
				memCache.Set(usagetype+storageClass+volumeType, usd, cache.NoExpiration)
				return usd
			}
		}
	}
	return usd
}

func getPricingFilters(bucketRegion string, objStorageClass string) (usageType, storageClass, volumeType string) {
	storageClass = storageTypes[objStorageClass].storageClass
	volumeType = storageTypes[objStorageClass].volumeType
	shortCode := storageTypes[objStorageClass].shortCode
	separator := "-"
	if shortCode == "" {
		separator = ""
	}
	usageType = regionCodes[bucketRegion] + "-TimedStorage" + separator + shortCode + separator + "-ByteHrs"
	return
}
