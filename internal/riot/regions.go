package riot

// Platform routing values (for summoner, league endpoints).
const (
	PlatformBR1  = "br1"  // Brazil
	PlatformEUN1 = "eun1" // EU Nordic & East
	PlatformEUW1 = "euw1" // EU West
	PlatformJP1  = "jp1"  // Japan
	PlatformKR   = "kr"   // Korea
	PlatformLA1  = "la1"  // Latin America North
	PlatformLA2  = "la2"  // Latin America South
	PlatformNA1  = "na1"  // North America
	PlatformOC1  = "oc1"  // Oceania
	PlatformTR1  = "tr1"  // Turkey
	PlatformRU   = "ru"   // Russia
	PlatformPH2  = "ph2"  // Philippines
	PlatformSG2  = "sg2"  // Singapore
	PlatformTH2  = "th2"  // Thailand
	PlatformTW2  = "tw2"  // Taiwan
	PlatformVN2  = "vn2"  // Vietnam
)

// Regional routing values (for match endpoints).
const (
	RegionAmericas = "americas"
	RegionAsia     = "asia"
	RegionEurope   = "europe"
	RegionSEA      = "sea"
)

// PlatformToRegion maps platform routing values to regional routing values.
var PlatformToRegion = map[string]string{
	PlatformNA1:  RegionAmericas,
	PlatformBR1:  RegionAmericas,
	PlatformLA1:  RegionAmericas,
	PlatformLA2:  RegionAmericas,
	PlatformKR:   RegionAsia,
	PlatformJP1:  RegionAsia,
	PlatformEUW1: RegionEurope,
	PlatformEUN1: RegionEurope,
	PlatformTR1:  RegionEurope,
	PlatformRU:   RegionEurope,
	PlatformOC1:  RegionSEA,
	PlatformPH2:  RegionSEA,
	PlatformSG2:  RegionSEA,
	PlatformTH2:  RegionSEA,
	PlatformTW2:  RegionSEA,
	PlatformVN2:  RegionSEA,
}
