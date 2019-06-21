# synth-snap
Studio sound maker

enable the driver

```
sudo vi /boot/uboot/config.txt

# comment out this line
dtparam=audio=on

# add this line
dtoverlay=hifiberry-dac

sudo reboot
```

install pulseaudio
```
snap install pulseaudio
```

install our snap

```
sudo snap install --edge codeverse-synth
sudo snap connect codeverse-synth:gpio-memory-control :gpio-memory-control
```

connect the plug
```
snap connect codeverse-synth:home :home
```

Enable the sound card
```
sudo /snap/codeverse-synth/current/bin/enable-sound-card
```

Now you can play a wav file by doing:
```
sudo pulseaudio-example /var/snap/pulseaudio-example/common/<file.wav>
```
