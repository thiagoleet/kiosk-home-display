package weather

func conditionFromCode(code int) WeatherCondition {
	switch code {
	case 0:
		return ConditionClear

	case 1, 2:
		return ConditionPartlyCloudy

	case 3:
		return ConditionOvercast

	case 45, 48:
		return ConditionFog

	case 51, 53, 55, 56, 57:
		return ConditionDrizzle

	case 61, 63, 65, 66, 67:
		return ConditionRain

	case 71, 73, 75, 77, 85, 86:
		return ConditionSnow

	case 80, 81, 82:
		return ConditionRainShowers

	case 95, 96, 99:
		return ConditionThunderstorm

	default:
		return ConditionClear
	}
}
