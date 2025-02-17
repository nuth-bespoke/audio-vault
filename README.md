![Audio Vault](/assets/logo.png?raw=true)

> noun: vault; a secure store where valuables are kept safe.

# Audio Vault

A web service which stores audio segments (WAV) from a Voice Recognition / Digital Dictation system.

Audio Vault's key features:

- Handles the safe storage of audio segments.
- Schedules the processing of the audio segments to normalise the Bit Rates.
- Concatenates the audio segments into a single audio recording for use by medical secretaries.
- Interfaces with the hospitals document management system to attach to the audio file to the letter.
- Handles the safe storage of audio orphans.
- Implements audio retention polices.


## Architectural Decisions

- The system will be developed using the [Go](https://go.dev/) programming language so that it can be cross compiled to Linux and Windows.
- The system will use an [SQLite](https://www.sqlite.org/) database to store its meta data.
- The system will offload audio processing to [SoX](https://linux.die.net/man/1/sox) Sound eXchange.


## Endpoints

A list of endpoints that the service will expose.
The table outlines which endpoints use logging and which require an API key (configured in the `settings.ini` file)

| endpoint        | logged | API Key | description                                                                  |
| --------------- |------- |---------|----------------------------------------------------------------------------- |
| /health-check   | N      | N       | Used by clients and automated checks to see if the service is running        |
| /record         | Y      | Y       | Allows clients to submit meta data about incoming audio segments             |
| /store          | Y      | Y       | Allows clients to submit audio segments for safe keeping                     |
| /orphan         | Y      | Y       | Allows clients to submit audio not associated with a specific letter         |
| /testing        | Y      | N       | A web user interface to allow tester to monitor the audio conversion process |
