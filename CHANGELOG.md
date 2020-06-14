# Changelog
All notable changes to this project will be documented in this file.

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased
 - Messaging: retrieve Reason for SMS delivery in case of failure, add GetMedia() function to retrieve the URL list
 - Add code to reconnect and handle reconnections

## [1.0.2] 2020-04-20
 - fix incoming Messaging (access Body)
 - Added ogg/opus decoding in the Tap example app.
 - Handle network timeouts at connect time.

## [1.0.1] 2019-12-20
 - all required settings retrieved through environment variables for the Example Apps.
 - support WSS (Secure Websocket) in the Tap example.
 - Add WebSocket Tap, extend Tap example app
 - Pass Detect timeout from the app down to RELAY command (all three detectors).

## [1.0.0] - 2019-11-25
- add RTP endpoint for Tap Example app. 
- fix unit test (run "go test -race" in CI)

## [1.0.0-rc.1] - 2019-11-13
- add Event Name and Event Payload properties per call, add getter for ControlID per Action
- add AMD and AMDAsync Methods
- decorate each Action with the Payload (Relay command params - can be inspected by the SDK user)
- add Play states for Prompt Action and run callbacks
- fix incorrect log level of basic logger
- remove silenced default logger code, upstream fix
- add callbacks OnAnswered, OnRinging, OnEnding, OnEnded, OnStateChange
- add properties "Failed" and "Type" for CallObj 

## [1.0.0-beta.2] - 2019-11-5
- expose Client 
- Use Environment variables in tests and examples.
- Use context driven timeout for HTTP client connection.
- Added Tasking API.
- Added Messaging API.
- Added Actions SendDigits, Tap, Prompt, Connect, with example apps.

## [1.0.0-beta.1] - 2019-10-16
### Added
- First release (beta.1)!

<!---
### Added
### Changed
### Removed
### Fixed
### Security
-->
