package service

import (
	"encoding/json"
	"fmt"

	"github.com/yukichan-bot-module/MiraiGo-module-weather/internal/pkg"
)

// CaiyunAPIUrl caiyun api url
const CaiyunAPIUrl string = "https://api.caiyunapp.com"

// CaiyunAPIVersion caiyun api version
const CaiyunAPIVersion string = "v2.6"

// Caiyun 彩云天气
// https://caiyunapp.com/
type Caiyun struct {
	APIKey string
}

// NewCaiyun Create Caiyun
func NewCaiyun(apiKey string) *Caiyun {
	return &Caiyun{
		APIKey: apiKey,
	}
}

// RealTime 实时天气情况
func (c *Caiyun) RealTime(longitude, latitude float64) (string, error) {
	url := fmt.Sprintf("%s/%s/%s/%f,%f/realtime", CaiyunAPIUrl, CaiyunAPIVersion, c.APIKey, longitude, latitude)
	var realtimeResponse CaiyunAPIRealTimeResponse
	responseBody, err := pkg.HTTPGetRequest(url, [][]string{
		{"unit", "metric:v2"},
		{"lang", "zh_CN"},
	})
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(responseBody, &realtimeResponse); err != nil {
		return "", err
	}
	if realtimeResponse.Status != "ok" {
		return "", fmt.Errorf("caiyun api error")
	}
	realTime := realtimeResponse.Result.RealTime
	if realTime.Status != "ok" {
		return "", fmt.Errorf("caiyun realtime error")
	}
	temperature := fmt.Sprintf("地表气温 %.1f ℃\n", realTime.Temperature)
	humidity := fmt.Sprintf("地表相对湿度 %.2f%%\n", realTime.Humidity*100)
	skycon := fmt.Sprintf("天气 %s\n", SkyconParse(realTime.Skycon))
	visibility := fmt.Sprintf("地表水平能见度 %.2f\n", realTime.Visibility)
	dswrf := fmt.Sprintf("向下短波辐射通量(W/M2) %.2f\n", realTime.Dswrf)
	windSpeed := fmt.Sprintf("当前风速 %.2f km/hr\n", realTime.Wind.Speed)
	windDirection := fmt.Sprintf("当前风向 %.2f° %s\n", realTime.Wind.Direction, windDirectionParse(realTime.Wind.Direction))
	pressure := fmt.Sprintf("地面气压 %.2f Pa\n", realTime.Pressure)
	apparentTemperature := fmt.Sprintf("体感温度 %.2f ℃\n", realTime.ApparentTemperature)
	intensity := fmt.Sprintf("本地降水强度 %.2f mm/hr\n", realTime.Precipitation.Local.Intensity)
	pm25 := fmt.Sprintf("PM25 浓度 %d μg/m3\n", realTime.AirQuality.PM25)
	pm10 := fmt.Sprintf("PM10 浓度 %d μg/m3\n", realTime.AirQuality.PM10)
	o3 := fmt.Sprintf("臭氧浓度 %d μg/m3\n", realTime.AirQuality.O3)
	so2 := fmt.Sprintf("二氧化硫浓度 %d μg/m3\n", realTime.AirQuality.SO2)
	no2 := fmt.Sprintf("二氧化氮浓度 %d μg/m3\n", realTime.AirQuality.NO2)
	co := fmt.Sprintf("一氧化碳浓度 %.2f μg/m3\n", realTime.AirQuality.CO)
	aqi := fmt.Sprintf("国标 AQI 指数 %d\n", realTime.AirQuality.AQI.CHN)
	aqiQuality := fmt.Sprintf("空气质量 %s\n", realTime.AirQuality.Description.CHN)
	airQuality := aqi + aqiQuality + pm25 + pm10 + o3 + so2 + no2 + co
	ultraviolet := fmt.Sprintf("紫外线强度 %s\n", realTime.LifeIndex.Ultraviolet.Description)
	comfort := fmt.Sprintf("舒适度 %s\n", realTime.LifeIndex.Comfortable.Description)
	origin := "信息来源：彩云天气"
	result := temperature + humidity + skycon + visibility + dswrf + windSpeed + windDirection +
		pressure + apparentTemperature + intensity + airQuality + ultraviolet + comfort + origin
	return result, nil
}

// Rain 短期内是否有雨
func (c *Caiyun) Rain(longitude, latitude float64) (string, error) {
	// TODO
	return "TODO", nil
}

// Tomorrow 明天的天气
func (c *Caiyun) Tomorrow(longitude, latitude float64) (string, error) {
	// TODO
	return "TODO", nil
}

// CaiyunAPIRealTimeResponse 实时天气情况返回
// https://docs.caiyunapp.com/docs/realtime
type CaiyunAPIRealTimeResponse struct {
	Status     string     `json:"status"`
	APIVersion string     `json:"api_version"`
	APIStatus  string     `json:"api_status"`
	Language   string     `json:"lang"`
	Unit       string     `json:"unit"`
	TZShift    int        `json:"tzshift"`
	Timezone   string     `json:"timezone"`
	ServerTime int64      `json:"server_time"`
	Location   [2]float64 `json:"location"`
	Result     struct {
		RealTime struct {
			Status      string  `json:"status"`
			Temperature float64 `json:"temperature"` // 地表 2 米气温
			Humidity    float64 `json:"humidity"`    // 地表 2 米湿度相对湿度(%)
			Cloudrate   float64 `json:"cloudrate"`   // 总云量(0.0-1.0)
			Skycon      string  `json:"skycon"`      // 天气现象
			Visibility  float64 `json:"visibility"`  // 地表水平能见度
			Dswrf       float64 `json:"dswrf"`       // 向下短波辐射通量(W/M2)
			Wind        struct {
				Speed     float64 `json:"speed"`     // 地表 10 米风速
				Direction float64 `json:"direction"` // 地表 10 米风向
			} `json:"wind"`
			Pressure            float64 `json:"pressure"`             // 地面气压
			ApparentTemperature float64 `json:"apparent_temperature"` // 体感温度
			Precipitation       struct {
				Local struct {
					Status     string  `json:"status"`
					DataSource string  `json:"datasource"`
					Intensity  float64 `json:"intensity"` // 本地降水强度
				} `json:"local"`
				Nearest struct {
					Status    string  `json:"status"`
					Distance  float64 `json:"distance"`  // 最近降水带与本地的距离
					Intensity float64 `json:"intensity"` // 最近降水处的降水强度
				} `json:"nearest"`
			} `json:"precipitation"`
			AirQuality struct {
				PM25 int     `json:"pm25"` // PM25 浓度(μg/m3)
				PM10 int     `json:"pm10"` // PM10 浓度(μg/m3)
				O3   int     `json:"o3"`   // 臭氧浓度(μg/m3)
				SO2  int     `json:"so2"`  // 二氧化硫浓度(μg/m3)
				NO2  int     `json:"no2"`  // 二氧化氮浓度(μg/m3)
				CO   float64 `json:"co"`   // 一氧化碳浓度(mg/m3)
				AQI  struct {
					CHN int `json:"chn"` // 国标 AQI
					USA int `json:"usa"`
				} `json:"aqi"`
				Description struct {
					CHN string `json:"chn"`
					USA string `json:"usa"`
				}
			} `json:"air_quality"`
			LifeIndex struct {
				Ultraviolet struct {
					Index       float64 `json:"index"`
					Description string  `json:"desc"`
				} `json:"ultraviolet"`
				Comfortable struct {
					Index       int    `json:"int"`
					Description string `json:"desc"`
				} `json:"comfort"`
			} `json:"life_index"` // 生活指数
		} `json:"realtime"`
		Primary int `json:"primary"`
	} `json:"result"`
}

type CaiyunAPIMinutelyResponse struct {
	Status     string    `json:"status"`
	APIVersion string    `json:"api_version"`
	APIStatus  string    `json:"api_status"`
	Lang       string    `json:"lang"`
	Unit       string    `json:"unit"`
	Tzshift    int       `json:"tzshift"`
	Timezone   string    `json:"timezone"`
	ServerTime int       `json:"server_time"`
	Location   []float64 `json:"location"`
	Result     struct {
		Minutely struct {
			Status          string    `json:"status"`
			Datasource      string    `json:"datasource"`
			Precipitation2H []float64 `json:"precipitation_2h"`
			Precipitation   []float64 `json:"precipitation"`
			Probability     []float64 `json:"probability"`
			Description     string    `json:"description"`
		} `json:"minutely"`
		Primary          int    `json:"primary"`
		ForecastKeypoint string `json:"forecast_keypoint"`
	} `json:"result"`
}

type CaiyunAPIHourlyResponse struct {
	Status     string    `json:"status"`
	APIVersion string    `json:"api_version"`
	APIStatus  string    `json:"api_status"`
	Lang       string    `json:"lang"`
	Unit       string    `json:"unit"`
	Tzshift    int       `json:"tzshift"`
	Timezone   string    `json:"timezone"`
	ServerTime int       `json:"server_time"`
	Location   []float64 `json:"location"`
	Result     struct {
		Hourly struct {
			Status        string `json:"status"`
			Description   string `json:"description"`
			Precipitation []struct {
				Datetime    string  `json:"datetime"`
				Value       float64 `json:"value"`
				Probability int     `json:"probability"`
			} `json:"precipitation"`
			Temperature []struct {
				Datetime string  `json:"datetime"`
				Value    float64 `json:"value"`
			} `json:"temperature"`
			ApparentTemperature []struct {
				Datetime string  `json:"datetime"`
				Value    float64 `json:"value"`
			} `json:"apparent_temperature"`
			Wind []struct {
				Datetime  string  `json:"datetime"`
				Speed     float64 `json:"speed"`
				Direction float64 `json:"direction"`
			} `json:"wind"`
			Humidity []struct {
				Datetime string  `json:"datetime"`
				Value    float64 `json:"value"`
			} `json:"humidity"`
			Cloudrate []struct {
				Datetime string  `json:"datetime"`
				Value    float64 `json:"value"`
			} `json:"cloudrate"`
			Skycon []struct {
				Datetime string `json:"datetime"`
				Value    string `json:"value"`
			} `json:"skycon"`
			Pressure []struct {
				Datetime string  `json:"datetime"`
				Value    float64 `json:"value"`
			} `json:"pressure"`
			Visibility []struct {
				Datetime string  `json:"datetime"`
				Value    float64 `json:"value"`
			} `json:"visibility"`
			Dswrf []struct {
				Datetime string  `json:"datetime"`
				Value    float64 `json:"value"`
			} `json:"dswrf"`
			AirQuality struct {
				Aqi []struct {
					Datetime string `json:"datetime"`
					Value    struct {
						Chn int `json:"chn"`
						Usa int `json:"usa"`
					} `json:"value"`
				} `json:"aqi"`
				Pm25 []struct {
					Datetime string `json:"datetime"`
					Value    int    `json:"value"`
				} `json:"pm25"`
			} `json:"air_quality"`
		} `json:"hourly"`
		Primary          int    `json:"primary"`
		ForecastKeypoint string `json:"forecast_keypoint"`
	} `json:"result"`
}

type CaiyunAPIDayilyResponse struct {
	Status     string     `json:"status"`
	APIVersion string     `json:"api_version"`
	APIStatus  string     `json:"api_status"`
	Language   string     `json:"lang"`
	Unit       string     `json:"unit"`
	TZShift    int        `json:"tzshift"`
	Timezone   string     `json:"timezone"`
	ServerTime int64      `json:"server_time"`
	Location   [2]float64 `json:"location"`
	Result     struct {
		Daily struct {
			Status string `json:"status"`
			Astro  []struct {
				Date    string `json:"date"`
				Sunrise struct {
					Time string `json:"time"`
				} `json:"sunrise"`
				Sunset struct {
					Time string `json:"time"`
				} `json:"sunset"`
			} `json:"astro"` // 日出日落时间，当地时区的时刻 (tzshift 不作用在这个变量)
			Precipitation08H20H []struct {
				Date        string  `json:"date"`
				Max         float64 `json:"max"`
				Min         float64 `json:"min"`
				Avg         float64 `json:"avg"`
				Probability int     `json:"probability"`
			} `json:"precipitation_08h_20h"` // 白天降水数据
			Precipitation20H32H []struct {
				Date        string  `json:"date"`
				Max         float64 `json:"max"`
				Min         float64 `json:"min"`
				Avg         float64 `json:"avg"`
				Probability int     `json:"probability"`
			} `json:"precipitation_20h_32h"` // 夜晚降水数据
			Precipitation []struct {
				Date        string  `json:"date"`
				Max         float64 `json:"max"`
				Min         float64 `json:"min"`
				Avg         float64 `json:"avg"`
				Probability int     `json:"probability"`
			} `json:"precipitation"` // 降水数据
			Temperature []struct {
				Date string  `json:"date"`
				Max  float64 `json:"max"`
				Min  float64 `json:"min"`
				Avg  float64 `json:"avg"`
			} `json:"temperature"` // 全天地表 2 米气温
			Temperature08H20H []struct {
				Date string  `json:"date"`
				Max  float64 `json:"max"`
				Min  float64 `json:"min"`
				Avg  float64 `json:"avg"`
			} `json:"temperature_08h_20h"` // 白天地表 2 米气温
			Temperature20H32H []struct {
				Date string  `json:"date"`
				Max  float64 `json:"max"`
				Min  float64 `json:"min"`
				Avg  float64 `json:"avg"`
			} `json:"temperature_20h_32h"` // 夜晚地表 2 米气温
			Wind []struct {
				Date string `json:"date"`
				Max  struct {
					Speed     float64 `json:"speed"`
					Direction float64 `json:"direction"`
				} `json:"max"`
				Min struct {
					Speed     float64 `json:"speed"`
					Direction float64 `json:"direction"`
				} `json:"min"`
				Avg struct {
					Speed     float64 `json:"speed"`
					Direction float64 `json:"direction"`
				} `json:"avg"`
			} `json:"wind"` // 全天地表 10 米风速
			Wind08H20H []struct {
				Date string `json:"date"`
				Max  struct {
					Speed     float64 `json:"speed"`
					Direction float64 `json:"direction"`
				} `json:"max"`
				Min struct {
					Speed     float64 `json:"speed"`
					Direction float64 `json:"direction"`
				} `json:"min"`
				Avg struct {
					Speed     float64 `json:"speed"`
					Direction float64 `json:"direction"`
				} `json:"avg"`
			} `json:"wind_08h_20h"` // 白天地表 10 米风速
			Wind20H32H []struct {
				Date string `json:"date"`
				Max  struct {
					Speed     float64 `json:"speed"`
					Direction float64 `json:"direction"`
				} `json:"max"`
				Min struct {
					Speed     float64 `json:"speed"`
					Direction float64 `json:"direction"`
				} `json:"min"`
				Avg struct {
					Speed     float64 `json:"speed"`
					Direction float64 `json:"direction"`
				} `json:"avg"`
			} `json:"wind_20h_32h"` // 夜晚地表 10 米风速
			Humidity []struct {
				Date string  `json:"date"`
				Max  float64 `json:"max"`
				Min  float64 `json:"min"`
				Avg  float64 `json:"avg"`
			} `json:"humidity"` // 地表 2 米相对湿度(%)
			Cloudrate []struct {
				Date string  `json:"date"`
				Max  float64 `json:"max"`
				Min  float64 `json:"min"`
				Avg  float64 `json:"avg"`
			} `json:"cloudrate"` // 云量(0.0-1.0)
			Pressure []struct {
				Date string  `json:"date"`
				Max  float64 `json:"max"`
				Min  float64 `json:"min"`
				Avg  float64 `json:"avg"`
			} `json:"pressure"` // 地面气压
			Visibility []struct {
				Date string  `json:"date"`
				Max  float64 `json:"max"`
				Min  float64 `json:"min"`
				Avg  float64 `json:"avg"`
			} `json:"visibility"` // 地表水平能见度
			Dswrf []struct {
				Date string  `json:"date"`
				Max  float64 `json:"max"`
				Min  float64 `json:"min"`
				Avg  float64 `json:"avg"`
			} `json:"dswrf"` // 向下短波辐射通量(W/M2)
			AirQuality struct {
				Aqi []struct {
					Date string `json:"date"`
					Max  struct {
						Chn int `json:"chn"`
						Usa int `json:"usa"`
					} `json:"max"`
					Avg struct {
						Chn int `json:"chn"`
						Usa int `json:"usa"`
					} `json:"avg"`
					Min struct {
						Chn int `json:"chn"`
						Usa int `json:"usa"`
					} `json:"min"`
				} `json:"aqi"` // 国标 AQI
				Pm25 []struct {
					Date string `json:"date"`
					Max  int    `json:"max"`
					Avg  int    `json:"avg"`
					Min  int    `json:"min"`
				} `json:"pm25"` // PM2.5 浓度(μg/m3)
			} `json:"air_quality"`
			Skycon []struct {
				Date  string `json:"date"`
				Value string `json:"value"`
			} `json:"skycon"` // 全天主要 天气现象
			Skycon08H20H []struct {
				Date  string `json:"date"`
				Value string `json:"value"`
			} `json:"skycon_08h_20h"` // 白天主要 天气现象
			Skycon20H32H []struct {
				Date  string `json:"date"`
				Value string `json:"value"`
			} `json:"skycon_20h_32h"` // 夜晚主要 天气现象
			LifeIndex struct {
				Ultraviolet []struct {
					Date        string `json:"date"`
					Index       string `json:"index"`
					Description string `json:"desc"` // 紫外线指数自然语言
				} `json:"ultraviolet"`
				CarWashing []struct {
					Date        string `json:"date"`
					Index       string `json:"index"`
					Description string `json:"desc"` // 洗车指数自然语言
				} `json:"carWashing"`
				Dressing []struct {
					Date        string `json:"date"`
					Index       string `json:"index"`
					Description string `json:"desc"` // 穿衣指数自然语言
				} `json:"dressing"`
				Comfort []struct {
					Date        string `json:"date"`
					Index       string `json:"index"`
					Description string `json:"desc"` // 舒适度指数自然语言
				} `json:"comfort"`
				ColdRisk []struct {
					Date        string `json:"date"`
					Index       string `json:"index"`
					Description string `json:"desc"` // 感冒指数自然语言
				} `json:"coldRisk"`
			} `json:"life_index"`
		} `json:"daily"`
		Primary int `json:"primary"`
	} `json:"result"`
}

type CaiyunAPIAlertResponse struct {
	Status     string     `json:"status"`
	APIVersion string     `json:"api_version"`
	APIStatus  string     `json:"api_status"`
	Lang       string     `json:"lang"`
	Unit       string     `json:"unit"`
	Tzshift    int        `json:"tzshift"`
	Timezone   string     `json:"timezone"`
	ServerTime int        `json:"server_time"`
	Location   [2]float64 `json:"location"`
	Result     struct {
		Alert struct {
			Status  string `json:"status"`
			Content []struct {
				Province      string     `json:"province"`    // 省，如"福建省"
				Status        string     `json:"status"`      // 预警信息的状态，如"预警中"
				Code          string     `json:"code"`        // 预警代码，如"0902"
				Description   string     `json:"description"` // 描述，如"三明市气象台 2020 年 04 月 21 日 12 时 19 分继续发布雷电黄色预警信号：预计未来 6 小时我市有雷电活动，局地伴有短时强降水、6-8 级雷雨大风等强对流天气。请注意防范！"
				RegionID      string     `json:"regionId"`
				County        string     `json:"county"`       // 县，如"无"
				Pubtimestamp  int        `json:"pubtimestamp"` // 发布时间，单位是 Unix 时间戳，如 1587443583
				Latlon        [2]float64 `json:"latlon"`
				City          string     `json:"city"`     // 市，如"三明市"
				AlertID       string     `json:"alertId"`  // 预警 ID，如 "35040041600001_20200421123203"
				Title         string     `json:"title"`    // 标题，如"三明市气象台发布雷电黄色预警[Ⅲ 级/较重]",
				Adcode        string     `json:"adcode"`   // 区域代码，如 "350400"
				Source        string     `json:"source"`   // 发布单位，如"国家预警信息发布中心",
				Location      string     `json:"location"` // 位置，如"福建省三明市"
				RequestStatus string     `json:"request_status"`
			} `json:"content"`
			Adcodes []struct {
				Adcode int    `json:"adcode"`
				Name   string `json:"name"`
			} `json:"adcodes"` // 行政区划层级信息
		} `json:"alert"`
		Primary int `json:"primary"`
	} `json:"result"`
}

// SkyconParse 天气现象解析
func SkyconParse(skycon string) string {
	var result string
	switch skycon {
	case "CLEAR_DAY":
		result = "晴（白天）"
	case "CLEAR_NIGHT":
		result = "晴（夜间）"
	case "PARTLY_CLOUDY_DAY":
		result = "多云（白天）"
	case "PARTLY_CLOUDY_NIGHT":
		result = "多云（夜间）"
	case "CLOUDY":
		result = "阴"
	case "LIGHT_HAZE":
		result = "轻度雾霾"
	case "MODERATE_HAZE":
		result = "中度雾霾"
	case "HEAVY_HAZE":
		result = "重度雾霾"
	case "LIGHT_RAIN":
		result = "小雨"
	case "MODERATE_RAIN":
		result = "中雨"
	case "HEAVY_RAINcc":
		result = "大雨"
	case "STORM_RAIN":
		result = "暴雨"
	case "FOG":
		result = "雾"
	case "LIGHT_SNOW":
		result = "小雪"
	case "MODERATE_SNOW":
		result = "中雪"
	case "HEAVY_SNOW":
		result = "大雪"
	case "STORM_SNOW":
		result = "暴雪"
	case "DUST":
		result = "浮尘"
	case "SAND":
		result = "沙尘"
	case "WIND":
		result = "大风"
	}
	return result
}

func windDirectionParse(direction float64) string {
	var result string
	switch v := direction; {
	case v >= 348.76 || v < 11.25:
		result = "北"
	case v >= 11.26 && v <= 33.75:
		result = "北东北"
	case v >= 33.76 && v <= 56.25:
		result = "东北"
	case v >= 56.26 && v <= 78.75:
		result = "东东北"
	case v >= 78.76 && v <= 101.25:
		result = "东"
	case v >= 101.26 && v <= 123.75:
		result = "东东南"
	case v >= 123.76 && v <= 146.25:
		result = "东南"
	case v >= 146.26 && v <= 168.75:
		result = "南东南"
	case v >= 168.76 && v <= 191.25:
		result = "南"
	case v >= 191.26 && v <= 213.75:
		result = "南西南"
	case v >= 213.76 && v <= 236.25:
		result = "西南"
	case v >= 236.26 && v <= 258.75:
		result = "西西南"
	case v >= 258.76 && v <= 281.25:
		result = "西"
	case v >= 281.26 && v <= 303.75:
		result = "西西北"
	case v >= 303.76 && v <= 326.25:
		result = "西北"
	case v >= 326.26 && v <= 348.75:
		result = "北西北"
	default:
		result = "error"
	}
	return result
}
