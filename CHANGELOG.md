# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## 1.1.0 - 2022-04-29
### Added
- new config opt (streamer user-agent)
- error msgs edited

## 1.0.0 - 2022-04-26 - big refactoring & breaking changes
### Changed
- almost all flags/cmd arguments moved to config file
- build scripts
### Added
- config file
- logging options (destination/verbosity)
- url extractor settings moved entirely to config file

## 0.7.0 - 2021-08-14
### Added
- m4a audio support (vf=m4a option)

## 0.6.0 - 2020-08-23
### Added
- error-headers option
- ignore-missing-headers option
- ignore-ssl-errors option

## 0.5.0 - 2020-08-11
### Added
- Clearer debug
- Flags (CLI options): debug, error video file, port, version, extractor select
- Custom url extractor
### Removed
- Port passing as argument

## 0.4.0 - 2020-08-03
Technical update

## 0.3.0 - 2017-11-03
### Added
- max video height and video format request

## 0.2.0 - 2017-10-31
### Added
- Content-Range header (fixed some seek errors)
- Show "corruped" video on errors

## 0.1.0 - 2017-09-31
### Added
- CHANGELOG
- README

### Removed

### Changed
