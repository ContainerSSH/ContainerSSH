package log

type Config struct {
	level Level
}

func NewConfig(
	level LevelString,
) (Config, error) {
	newLevel, err := LevelFromString(level)
	if err != nil {
		return Config{}, err
	}

	return Config{
		level:    newLevel,
	}, nil
}

func (config * Config) GetLevel() Level {
	return config.level
}

func (config * Config) GetLevelName() LevelString {
	levelName, err := config.level.ToString()
	if err != nil {
		//This should never happen
		panic(err)
	}
	return levelName
}

func (config * Config) WithLevel(level Level) (*Config, error) {
	err := level.Validate()
	if err != nil {
		return nil, err
	}
	return &Config{
		level:    level,
	}, nil
}

func (config * Config) WithLevelName(levelName LevelString) (*Config, error) {
	level, err := LevelFromString(levelName)
	if err != nil {
		return nil, err
	}

	return &Config{
		level:    level,
	}, nil
}
