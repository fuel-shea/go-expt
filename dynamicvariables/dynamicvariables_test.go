package dynamicvariables_test

import (
	"dynamic-variables-server/dynamicvariables"
	"gopkg.in/mgo.v2"
	"testing"
)

var gameID = "gid"

func TestVarsFromFeatures_equals(t *testing.T) {
	mgoDBHost := "localhost"
	mgoDBName := "dynamicvariables_test"

	for _, tc := range testCases {
		err := populateDB(mgoDBHost, mgoDBName, tc)
		if err != nil {
			t.Fatal(err)
		}

		dvFactory, err := dynamicvariables.NewDynoVarFactory(mgoDBHost, mgoDBName)
		if err != nil {
			t.Fatal(err)
		}
		dvSource := dvFactory.NewDynoVarSource()

		varsResult, err := dvSource.VarsFromFeatures(tc.query, gameID)
		if err != nil {
			t.Fatal(err)
		}

		if varsResult["whammyChance"] != tc.expected["whammyChance"] {
			t.Error("Expected 'whammyChance' to be %q, but it was %v", tc.expected["whammyChance"], varsResult["whammyChance"])
		}
		if varsResult["randomMax"] != tc.expected["randomMax"] {
			t.Error("Expected 'randomMax' to be %q, but it was %v", tc.expected["randomMax"], varsResult["randomMax"])
		}
	}
}

type testCase struct {
	gameRuleData GameRuleData
	features     [][]Feature
	variables    []Variable
	query        map[string]interface{}
	expected     map[string]interface{}
}

type GameRuleData struct {
	GameID        string   `bson:"game_id"`
	VariableTypes []string `bson:"variable_types"`
	FeatureTypes  []string `bson:"feature_types"`
	NumRules      float64  `bson:"num_rules"`
}

type Feature struct {
	GameID  string `bson:"game_id"`
	RuleIdx int    `bson:"rule_idx"`
	Type    string `bson:"type"`
	Val     string `bson:"value"`
	Mod     string `bson:"mod"`
}

type Variable struct {
	GameID       string  `bson:"game_id"`
	RuleIdx      float64 `bson:"rule_idx"`
	RandomMax    string  `bson:"randomMax"`
	WhammyChance string  `bson:"whammyChance"`
}

var testCases = []testCase{
	{
		gameRuleData: GameRuleData{
			NumRules: float64(6),
		},
		features: [][]Feature{
			{
				Feature{Val: "CA", Mod: "="},
				Feature{Val: "iOS", Mod: "="},
				Feature{Val: "M", Mod: "="},
			},
			{
				Feature{Val: "US", Mod: "="},
				Feature{Val: "iOS", Mod: "="},
				Feature{Val: "M", Mod: "="},
			},
			{
				Feature{Val: "JP", Mod: "="},
				Feature{Val: "Android", Mod: "="},
				Feature{Val: "any", Mod: "="},
			},
			{
				Feature{Val: "CA", Mod: "="},
				Feature{Val: "Android", Mod: "="},
				Feature{Val: "F", Mod: "="},
			},
			{
				Feature{Val: "JP", Mod: "="},
				Feature{Val: "iOS", Mod: "="},
				Feature{Val: "any", Mod: "="},
			},
			{
				Feature{Val: "JP", Mod: "="},
				Feature{Val: "iOS", Mod: "="},
				Feature{Val: "F", Mod: "="},
			},
			{
				Feature{Val: "any", Mod: "="},
				Feature{Val: "any", Mod: "="},
				Feature{Val: "any", Mod: "="},
			},
		},
		variables: []Variable{
			Variable{RandomMax: "0", WhammyChance: "0"},
			Variable{RandomMax: "1", WhammyChance: "1"},
			Variable{RandomMax: "2", WhammyChance: "2"},
			Variable{RandomMax: "3", WhammyChance: "3"},
			Variable{RandomMax: "4", WhammyChance: "4"},
			Variable{RandomMax: "5", WhammyChance: "5"},
			Variable{RandomMax: "6", WhammyChance: "6"},
		},
		query: map[string]interface{}{
			"Country": "JP",
			"Device":  "iOS",
			"Gender":  "F",
		},
		expected: map[string]interface{}{
			"randomMax":    "4",
			"whammyChance": "4",
		},
	},
}

func populateDB(mgoDBHost, mgoDBName string, tc testCase) error {
	tc.gameRuleData.GameID = gameID
	tc.gameRuleData.FeatureTypes = []string{"Country", "Device", "Gender"}
	tc.gameRuleData.VariableTypes = []string{"randomMax", "whammyChance"}

	mgoSess, err := mgo.Dial(mgoDBHost)
	if err != nil {
		return err
	}
	mgoDB := mgoSess.DB(mgoDBName)
	err = mgoDB.DropDatabase()
	if err != nil {
		return err
	}

	err = mgoDB.C("game_rule_data").Insert(tc.gameRuleData)
	if err != nil {
		return err
	}

	featuresColl := mgoDB.C("features")
	variablesColl := mgoDB.C("variables")
	for ruleIdx, _ := range tc.features {
		variable := tc.variables[ruleIdx]
		variable.GameID = gameID
		variable.RuleIdx = float64(ruleIdx)
		err := variablesColl.Insert(variable)
		if err != nil {
			return err
		}

		for featIdx, feat := range tc.features[ruleIdx] {
			feat.GameID = gameID
			feat.RuleIdx = ruleIdx
			feat.Type = tc.gameRuleData.FeatureTypes[featIdx]
			err := featuresColl.Insert(feat)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
