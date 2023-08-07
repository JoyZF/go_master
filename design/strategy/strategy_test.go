package strategy

import (
	"fmt"
	"testing"
)

// City 城市
type City struct {
	name    string
	feature string
	season  Season
}

// NewCity 根据名称及季候特征创建城市
func NewCity(name, feature string) *City {
	return &City{
		name:    name,
		feature: feature,
	}
}

// SetSeason 设置不同季节，类似天气在不同季节的不同策略
func (c *City) SetSeason(season Season) {
	c.season = season
}

// String 显示城市的气候信息
func (c *City) String() string {
	return fmt.Sprintf("%s%s，%s", c.name, c.feature, c.season.ShowWeather(c.name))
}

func TestStrategy(t *testing.T) {
	NewCity("北京", "首都").SetSeason(NewSpring())
}
