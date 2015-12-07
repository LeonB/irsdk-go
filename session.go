package irsdk

import yaml "gopkg.in/yaml.v2"

type SessionData struct {
	WeekendInfo   WeekendInfo   `yaml:"WeekendInfo"`
	SessionInfo   SessionInfo   `yaml:"SessionInfo"`
	CameraInfo    CameraInfo    `yaml:"CameraInfo"`
	RadioInfo     RadioInfo     `yaml:"RadioInfo"`
	DriverInfo    DriverInfo    `yaml:"DriverInfo"`
	SplitTimeInfo SplitTimeInfo `yaml:"SplitTimeInfo"`
}

// https://github.com/smyrman/units
// quantity + unit
type unit string

type WeekendInfo struct {
	// TrackName string -> TrackName string `yaml:"TrackName"`
	// vim: s/\(\t\(.\{-}\) \).*/\0 `yaml:"\2"`
	TrackName              string           `yaml:"TrackName"`
	TrackID                int              `yaml:"TrackID"`
	TrackLength            string           `yaml:"TrackLength"`
	TrackDisplayName       string           `yaml:"TrackDisplayName"`
	TrackDisplayShortName  string           `yaml:"TrackDisplayShortName"`
	TrackConfigName        string           `yaml:"TrackConfigName"`
	TrackCity              string           `yaml:"TrackCity"`
	TrackCountry           string           `yaml:"TrackCountry"`
	TrackAltitude          unit             `yaml:"TrackAltitude"`
	TrackLatitude          unit             `yaml:"TrackLatitude"`
	TrackLongitude         unit             `yaml:"TrackLongitude"`
	TrackNumTurns          int              `yaml:"TrackNumTurns"`
	TrackPitSpeedLimit     unit             `yaml:"TrackPitSpeedLimit"`
	TrackType              string           `yaml:"TrackType"`
	TrackWeatherType       string           `yaml:"TrackWeatherType"`
	TrackSkies             string           `yaml:"TrackSkies"`
	TrackSurfaceTemp       unit             `yaml:"TrackSurfaceTemp"`
	TrackAirTemp           unit             `yaml:"TrackAirTemp"`
	TrackAirPressure       unit             `yaml:"TrackAirPressure"`
	TrackWindVel           unit             `yaml:"TrackWindVel"`
	TrackWindDir           unit             `yaml:"TrackWindDir"`
	TrackRelativeHumidity  unit             `yaml:"TrackRelativeHumidity"`
	TrackFogLevel          unit             `yaml:"TrackFogLevel"`
	SeriesID               int              `yaml:"SeriesID"`
	SeasonID               int              `yaml:"SeasonID"`
	SessionID              int              `yaml:"SessionID"`
	SubSessionID           int              `yaml:"SubSessionID"`
	LeagueID               int              `yaml:"LeagueID"`
	Official               int              `yaml:"Official"`
	RaceWeek               int              `yaml:"RaceWeek"`
	EventType              string           `yaml:"EventType"`
	Category               string           `yaml:"Category"`
	SimMode                string           `yaml:"SimMode"`
	TeamRacing             int              `yaml:"TeamRacing"`
	MinDrivers             int              `yaml:"MinDrivers"`
	MaxDrivers             int              `yaml:"MaxDrivers"`
	DCRuleSet              string           `yaml:"DCRuleSet"`
	QualifierMustStartRace intToBool        `yaml:"QualifierMustStartRace"`
	NumCarClasses          int              `yaml:"NumCarClasses"`
	NumCarTypes            int              `yaml:"NumCarTypes"`
	WeekendOptions         WeekendOptions   `yaml:"WeekendOptions"`
	TelemetryOptions       TelemetryOptions `yaml:"TelemetryOptions"`
}

type WeekendOptions struct {
	NumStarters         int       `yaml:"NumStarters"`
	StartingGrid        string    `yaml:"StartingGrid"`
	QualifyScoring      string    `yaml:"QualifyScoring"`
	CourseCautions      string    `yaml:"CourseCautions"`
	StandingStart       intToBool `yaml:"StandingStart"`
	Restarts            string    `yaml:"Restarts"`
	WeatherType         string    `yaml:"WeatherType"`
	Skies               string    `yaml:"Skies"`
	WindDirection       unit      `yaml:"WindDirection"`
	WindSpeed           unit      `yaml:"WindSpeed"`
	WeatherTemp         unit      `yaml:"WeatherTemp"`
	RelativeHumidity    unit      `yaml:"RelativeHumidity"`
	FogLevel            unit      `yaml:"FogLevel"`
	Unofficial          intToBool `yaml:"Unofficial"`
	CommercialMode      string    `yaml:"CommercialMode"`
	NightMode           intToBool `yaml:"NightMode"`
	IsFixedSetup        intToBool `yaml:"IsFixedSetup"`
	StrictLapsChecking  string    `yaml:"StrictLapsChecking"`
	HasOpenRegistration intToBool `yaml:"HasOpenRegistration"`
	HardcoreLevel       int       `yaml:"HardcoreLevel"`
}

type TelemetryOptions struct {
	TelemetryDiskFile string `yaml:"TelemetryDiskFile"`
}

type SessionInfo struct {
	Sessions []Session `yaml:"Sessions"`
}

type Session struct {
	SessionNum             int                `yaml:"SessionNum"`
	SessionLaps            string             `yaml:"SessionLaps"`
	SessionTime            string             `yaml:"SessionTime"`
	SessionNumLapsToAvg    int                `yaml:"SessionNumLapsToAvg"`
	SessionType            string             `yaml:"SessionType"`
	ResultsPositions       []ResultPosition   `yaml:"ResultsPositions"`
	ResultsFastestLap      []ResultFastestLap `yaml:"ResultsFastestLap"`
	ResultsAverageLapTime  float32            `yaml:"ResultsAverageLapTime"`
	ResultsNumCautionFlags int                `yaml:"ResultsNumCautionFlags"`
	ResultsNumCautionLaps  int                `yaml:"ResultsNumCautionLaps"`
	ResultsNumLeadChanges  int                `yaml:"ResultsNumLeadChanges"`
	ResultsLapsComplete    int                `yaml:"ResultsLapsComplete"`
	ResultsOfficial        int                `yaml:"ResultsOfficial"`
}

type ResultPosition struct {
	CarIdx        int     `yaml:"CarIdx"`
	Position      int     `yaml:"Position"`
	ClassPosition int     `yaml:"ClassPosition"`
	FastestTime   float32 `yaml:"FastestTime"`
	Lap           int     `yaml:"Lap"`
	LastTime      float32 `yaml:"LastTime"`
	LapsComplete  int     `yaml:"LapsComplete"`
	LapsDriven    int     `yaml:"LapsDriven"`
	ReasonOutId   int     `yaml:"ReasonOutId"`
}

type ResultFastestLap struct {
	CarIdx      int     `yaml:"CarIdx"`
	FastestLap  int     `yaml:"FastestLap"`
	FastestTime float32 `yaml:"FastestTime"`
}

type CameraInfo struct {
	Groups []CameraGroup `yaml:"Groups"`
}

type CameraGroup struct {
	GroupNum  int      `yaml:"GroupNum"`
	GroupName string   `yaml:"GroupName"`
	IsScenic  bool     `yaml:"IsSenic"`
	Cameras   []Camera `yaml:"Cameras"`
}

type Camera struct {
	CameraNum  int    `yaml:"CameraNum"`
	CameraName string `yaml:"CameraName"`
}

type RadioInfo struct {
	SelectedRadioNum int     `yaml:"SelectedRadioNum"`
	Radios           []Radio `yaml:"Radios"`
}

type Radio struct {
	RadioNum            int         `yaml:"RadioNum"`
	HopCount            int         `yaml:"HopCount"`
	NumFrequencies      int         `yaml:"NumFrequencies"`
	TunedToFrequencyNum int         `yaml:"TunedToFrequencyNum"`
	ScanningIsOn        intToBool   `yaml:"ScanningIsOn"`
	Frequencies         []Frequency `yaml:"Frequencies"`
}

type Frequency struct {
	FrequencyNum  int       `yaml:"FrequencyNum"`
	FrequencyName string    `yaml:"FrequencyName"`
	Priority      int       `yaml:"Priority"`
	CarIdx        int       `yaml:"CarIdx"`
	EntryIdx      int       `yaml:"EntryIdx"`
	ClubID        int       `yaml:"ClubID"`
	CanScan       intToBool `yaml:"CanScan"`
	CanSquawk     intToBool `yaml:"CanSquawk"`
	Muted         intToBool `yaml:"Muted"`
	IsMutable     intToBool `yaml:"IsMutable"`
	IsDeletable   intToBool `yaml:"IsDeletable"`
}

type DriverInfo struct {
	DriverCarIdx          int      `yaml:"DriverCarIdx"`
	DriverHeadPosX        float32  `yaml:"DriverHeadPosX"`
	DriverHeadPosY        float32  `yaml:"DriverHeadPosY"`
	DriverHeadPosZ        float32  `yaml:"DriverHeadPosZ"`
	DriverCarIdleRPM      float32  `yaml:"DriverCarIdleRPM"`
	DriverCarRedLine      float32  `yaml:"DriverCarRedLine"`
	DriverCarFuelKgPerLtr float32  `yaml:"DriverCarFuelKgPerLtr"`
	DriverCarSLFirstRPM   float32  `yaml:"DriverCarSLFirstRPM"`
	DriverCarSLShiftRPM   float32  `yaml:"DriverCarSLShiftRPM"`
	DriverCarSLLastRPM    float32  `yaml:"DriverCarSLLastRPM"`
	DriverCarSLBlinkRPM   float32  `yaml:"DriverCarSLBlinkRPM"`
	DriverPitTrkPct       float32  `yaml:"DriverPitTrkPct"`
	Drivers               []Driver `yaml:"Drivers"`
}

type Driver struct {
	CarIdx     int    `yaml:"CarIdx"`
	UserName   string `yaml:"UserName"`
	AbbrevName string `yaml:"AbbrevName"`
	Initials   string `yaml:"Initials"`
	UserID     int    `yaml:"UserID"`
	TeamID     int    `yaml:"TeamID"`
	TeamName   string `yaml:"TeamName"`
	// Or shoud CarNumber be an int?
	CarNumber             string `yaml:"CarNumber"`
	CarNumberRaw          int    `yaml:"CarNumberRaw"`
	CarPath               string `yaml:"CarPath"`
	CarClassID            int    `yaml:"CarClassID"`
	CarID                 int    `yaml:"CarID"`
	CarScreenName         string `yaml:"CarScreenName"`
	CarScreenNameShort    string `yaml:"CarScreenNameShort"`
	CarClassShortName     string `yaml:"CarClassShortName"`
	CarClassRelSpeed      int    `yaml:"CarClassRelSpeed"`
	CarClassLicenseLevel  int    `yaml:"CarClassLicenseLevel"`
	CarClassMaxFuel       unit   `yaml:"CarClassMaxFuel"`
	CarClassWeightPenalty unit   `yaml:"CarClassWeightPenalty"`
	// CarClassColor: 0xffffff
	IRating     int    `yaml:"IRating"`
	LicLevel    int    `yaml:"LicLevel"`
	LicSubLevel int    `yaml:"LicSubLevel"`
	LicString   string `yaml:"LicString"`
	// LicColor: 0xfc8a27
	IsSpectator intToBool `yaml:"IsSpectator"`
	// CarDesignStr: 0,FFFFFF,ED2129,2A3795
	// HelmetDesignStr: 56,000000,000000,000000
	// SuitDesignStr: 0,000000,000000,000000
	// CarNumberDesignStr: 0,0,FFFFFF,777777,000000
	CarSponsor_1 int `yaml:"CarSponsor_1"`
	CarSponsor_2 int `yaml:"CarSponsor_2"`
}

type SplitTimeInfo struct {
	Sectors []Sector `yaml:"Sectors"`
}

type Sector struct {
	SectorNum      int     `yaml:"SectorNum"`
	SectorStartPct float32 `yaml:"SectorStartPct"`
}

type intToBool bool

func (i *intToBool) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var intResult int
	err := unmarshal(&intResult)
	if err != nil {
		return err
	}

	if intResult > 0 {
		*i = true
	} else {
		*i = false
	}

	return nil
}

// @TODO: should this accept an io.Reader?
func NewSessionDataFromBytes(yamlData []byte) (*SessionData, error) {
	sessionData := newSessionData()

	// Convert yaml to struct
	err := yaml.Unmarshal(yamlData, &sessionData)
	if err != nil || sessionData == nil {
		return nil, err
	}

	return sessionData, nil
}

func newSessionData() *SessionData {
	return &SessionData{}
}
