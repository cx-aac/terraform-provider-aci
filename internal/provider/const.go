package provider

import "time"

// Retry defaults
const Factor = 3
const MinDelay = 4 * time.Second
const MaxDelay = 60 * time.Second

// List of attributes to be not stored in state
var IgnoreAttr = []string{"extMngdBy", "lcOwn", "modTs", "monPolDn", "uid", "dn", "rn", "configQual", "configSt", "virtualIp"}

// List of attributes to be only written to state from config
var WriteOnlyAttr = []string{"childAction"}

// List of classes where 'rsp-prop-include=config-only' does not return the desired objects/properties
var FullClasses = []string{"firmwareFwGrp", "maintMaintGrp", "maintMaintP", "firmwareFwP"}
