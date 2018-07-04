# Go Ethereum - Korean 

이더리움 한글 주석 프로젝트입니다.

v0.2 현재 126개 파일에 약 1000줄의 주석이 한글로 변환되었습니다. 

동작 분석 결과는 https://steemit.com/@sigmoid를 통해 업데이트 중입니다.


## Packages in folder.
### Accounts: 계정과 키스토어, 계약계정, 지갑관련 기능
 * package accounts
 * package abi 
 * package bind
 * package backends
 * package keystore
 * package usbwallet
 * package trezor

### bmt
 * package bmt: 바이너리 머클 트리 

### cmd: geth 포함, go-ethereum에서 지원하는 실행 가능한 바이너리들
 * package utils
 * package browser
 * packages - Main: Executable binaries

### common: solidity 컴파일러 등 공통으로 쓰이는 패키지들
 * package compiler
 * package common
 * package bitutil, fdlimit, hexutil, math, mclock, number

### consensus: 합의구현부: POA, POW 마이닝, DAO & Pork
 * package consensus
 * package clique
 * package ethash
 * package misc

### console
 * package console

### contract: 스마트 계약관련
 * package chequebook
 * package contract
 * package main
 * package ens

### core: 이더리움 코어. 상태, DB, EVM
 * package asm
 * package core
 * package bloombits
 * package main
 * package rawdb
 * package state
 * package types
 * package runtime
 * package vm

### crypto: 암호화 관련
 * package bn256
 * package crypto
 * package ecies
 * package randentropy
 * package secp256k1
 * package sha3

### dashboard: 이더리움 대시보드
 * package dashboard

### eth: 이더리움 프로토콜
 * package eth
 * package downloader
 * package fetcher
 * package filters
 * package gasprice
 * package tracers

### ethclient
 * package ethclient

### ethdb
 * package ethdb

### ethstats
 * package ethstats

### event
 * package event
 * package filter

### internal 
 * package build
 * package cmdtest
 * package debug
 * package ethapi
 * package guide
 * package jsre
 * package web3ext

### les
 * package les
 * package flowcontrol

### light
 * package light

### log
 * package log
 * package term

### metrics
 * package exp
 * package influxdb
 * package librato
 * package main

### mobile
 * package geth

### node: 이더리움 노드
 * package node

### p2p: p2p 관련
 * package discover
 * package discv5
 * package enr
 * package nat
 * package netutil
 * package protocols
 * package p2p
 * package adapters
 * package main
 * package simulations

### params
 * package params

### rlp
 * package rlp

### rpc
 * package rpc

### signer
 * package deps
 * package rules
 * package storage

### swarm
 * package api
 * package client
 * package http
 * package fuse
 * package metrics
 * package network
 * package kademlia
 * package swap
 * package swarm

### trie
 * package trie

### whisper
 * package mailserver
 * package shhclient
 * package whisperv5
 * package whisperv6

### vendor
 * package fs
 autorest
 azure
 date
 color
 memsize
 memsizeui
 termui
 ole
 oleutil
 stack
 proto
 descriptor
 lru
 simplelru
 internetgateway1
 goupnp
 httpu
 scpd
 soap
 models
 escape
 natpmp
 httprouter
 hid
 colorable
 isatty
 runewidth
 wordwrap
 stringutil
 ast
 toml
 termbox
 tablewriter
 uuid
 liner
 errors
 difflib
 flock
 notify
 otto
 dbg
 file
 parser
 registry
 token
 cors
 xhandler
 assert
 require
 cache
 comparer
 iterator
 journal
 memdb
 opt
 table
 util
 leveldb
 cast5
 curve25519
 ed25519
 edwards25519
 armor
 openpgp
 elgamal
 packet
 s2k 
 pbkdf2 
 ripemd160 
 scrypt
 terminal
 ssh
 context 
 atom
 charset 
 html 
 websocket
 syncmap
 unix
 windows
 charmap
 encoding 
 htmlindex
 identifier
 internal
 japanese
 korean
 simplifiedchinese
 traditionalchinese
 unicode
 tag 
 utf8internal
 language
 transform 
 astutil
 imports
 cli


## GETH block
![](https://steemitimages.com/0x0/https://cdn.steemitimages.com/DQmUYdekGgmf7jKE4T3dhuXrsofMJHLyTK6F92XDZbXwVT4/image.png) 

## GETH function flow

![](https://steemitimages.com/0x0/https://cdn.steemitimages.com/DQmSS9KhboUC13eutXRdD8cruNJqAsfVw6Htb5HhCvMg5s5/image.png)

![](https://steemitimages.com/0x0/https://cdn.steemitimages.com/DQmPrGtgbxg8fvAV8oWneto4sZhfysD5gWfyCsBnBn71vCj/image.png)

![](https://steemitimages.com/0x0/https://cdn.steemitimages.com/DQmQV9QmVCQ5zZNkfxrY4VQPoSwJKCsiyxXM9aEoKtVADE5/image.png)

![](https://steemitimages.com/0x0/https://cdn.steemitimages.com/DQmTViEMDGUaXaG51UG7ze2tV3RDpgRhEUcPjCsW1qmjBS8/image.png)

![](https://steemitimages.com/0x0/https://cdn.steemitimages.com/DQmUioaCerYaU7VPDdyFZFRCQHioshJ9HzWM8WsmzTeagLF/image.png)

![](https://steemitimages.com/0x0/https://cdn.steemitimages.com/DQmbJ5UxC4shoWg4DjNpqYDYdAHDTgHPcVsnsaioF4S5oB9/image.png)

![](https://steemitimages.com/0x0/https://cdn.steemitimages.com/DQmVfznwXtZir1izJf8uYQCUs25DBNvc7VJSPYoYLMc24Nx/image.png)

![](https://steemitimages.com/0x0/https://cdn.steemitimages.com/DQmebE7oGmGupQazu2LxdqaCDtZAWamskGzNVzmnqJ7CPAE/image.png)

![](https://steemitimages.com/0x0/https://cdn.steemitimages.com/DQmWv4Nadia8oywxc3Pw2DCuihFYwJuJCge9kiwRat95C6c/image.png)

![](https://cdn.steemitimages.com/DQmT6WHwvjgxbL9qasRB4R64XzQug924P1GPz1HdoBrZGZc/image.png)

![](https://cdn.steemitimages.com/DQmPge1ZTyTyGV6LCn7HX6Ar1jizptCRTHJ8L4tF3uJgsWi/image.png)

![](https://cdn.steemitimages.com/DQmRdzfLRzzan9LMMVQHeeUX8LVwXKJd7doB9GyJgk8HGuv/image.png)
