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

![Audio Vault](/assets/screenshot.png?raw=true)

After individual segments are stitched together the system generates a wav form image of the resulting audio dictation.

![Audio Vault](/assets/audiowaveform.png?raw=true)

Future versions will incorporate [peaks.js](https://github.com/bbc/peaks.js) to provide real-time wav form syncronisation with the audio playback.

## Architectural Decisions

- The system will be developed using the [Go](https://go.dev/) programming language so that it can be cross compiled to Linux and Windows.
- The system will be run from a primary server (node) and will use sync tools to keep a warm standby server (node) in a ready state.
- The web service should be run via a reverse proxy from a web server, e.g. IIS.
- The web service can be installed on Windows as a service using the [NSSM](https://nssm.cc/) tool.
- The system will use an [SQLite](https://www.sqlite.org/) database to store its meta data.
- The system will use [SQLite Rsync](https://www.sqlite.org/rsync.html) to backup the SQLite database to a warm standby server.
- The system will use [rClone](https://rclone.org/local/) to sync audio to a warm standby server.
- The system will offload audio processing to [SoX](https://linux.die.net/man/1/sox) Sound eXchange.
- The system will offload audio wav form generation to [audiowaveform](https://github.com/bbc/audiowaveform)


## Endpoints

A list of endpoints that the service will expose.
The table outlines which endpoints use logging and which require an API key (configured in the `settings.ini` file)

| endpoint            | API Key | description                                                                       |
| ------------------- |---------|---------------------------------------------------------------------------------- |
| /dashboard          | N       | A web user interface to allow tester to monitor the audio conversion process      |
| /dictation          | N       | A web user interface to retrieve the final dictation and the segments that made it|
| /health-check       | N       | Used by clients and automated checks to see if the service is running             |
| /orphan             | Y       | Allows clients to submit audio not associated with a specific letter              |
| /server-side-events | N       | Sever push of CPU usage and list of pending segments that are being processed     |
| /store              | Y       | Allows clients to submit audio segments for safe keeping                          |
| /stream             | N       | Allows web user interface to stream audio to a HTML5 audio element                |
| /user               | N       | A web user interface to retrieve logs for a given user from a turso database      |


## Running Web Service from IIS using a Reverse Proxy
 
To configure IIS with Reverse Proxy Request Routing support, you'll need to install
[application-request-routing](https://iis-umbraco.azurewebsites.net/downloads/microsoft/application-request-routing)
 
Then set-up a basic `web.config` file with a single `rewrite` rule to forward all requests/traffic to the web service.
Substitute the port number, if you've changed it from the default of `1969`.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<configuration>
    <system.webServer>
    <rewrite>
        <rules>
            <rule name="ReverseProxyAudioVault" enabled="true" stopProcessing="true">
                <match url="(.*)" ignoreCase="true" />
                <action type="Rewrite" url="http://localhost:1969/{R:1}" />
                <serverVariables>
                    <set name="HTTP_AUTHORIZATION" value="{HTTP_AUTHORIZATION}" />
                </serverVariables>
            </rule>
        </rules>
    </rewrite>
    <httpProtocol>
        <customHeaders>
            <remove name="X-Powered-By" />
        </customHeaders>
    </httpProtocol>
    </system.webServer>
</configuration>
```