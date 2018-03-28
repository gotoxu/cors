# assert [![Build Status](https://travis-ci.org/ycyz/assert.svg?branch=master)](https://travis-ci.org/ycyz/assert)
Package assert provides utilities for Go testing


## func Empty
`func Empty(t Fataler, arr interface{}, a ...interface{})`

Empty确保`arr`是一个空集合，`arr`必须是数组或者切片

## func NotEmpty
`func NotEmpty(t Fataler, arr interface{}, a ...interface{})`

NotEmpty确保`arr`是一个非空集合，`arr`必须是数组或切片

## func DeepEqual
`func DeepEqual(t Fataler, actual, expected interface{}, a ...interface{})`

DeepEqual确保`actual`和`expected`是相等的，使用`reflect.DeepEqual`对比

## func NotDeepEqual
`func NotDeepEqual(t Fataler, actual, expected interface{}, a ...interface{})`

NotDeepEqual确保`actual`和`expected`不相等，使用`reflect.DeepEuqal`对比

## func Error
`func Error(t Fataler, err error, re *regexp.Regexp, a ...interface{})`

Error确保给定的错误信息于正则表达式`re`匹配

## func StringContains
`func StringContains(t Fataler, s, substr string, a ...interface{})`

StringContains判断给定的字符串`s`中是否包含子串`substr`

## func StringDoesNotContain
`func StringDoesNotContain(t Fataler, s, substr string, a ...interface{})`

StringDoesNotContain判断给定的字符串`s`中是否未包含子串`substr`

## func Nil
`func Nil(t Fataler, v interface{}, a ...interface{})`

Nil判断给定的对象`v`是否为`nil`

## func NotNil
`func NotNil(t Fataler, v interface{}, a ...interface{})`

NotNil判断给定的对象`v`是否不是`nil`

## func True
`func True(t Fataler, v bool, a ...interface{})`

True判断给定的值`v`是否为真

## func False
`func False(t Fataler, v bool, a ...interface{})`

False判断给定的值`v`是否为假