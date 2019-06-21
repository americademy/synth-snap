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

install our snap

```
sudo snap install --edge codeverse-synth
sudo snap connect codeverse-synth:gpio-memory-control :gpio-memory-control

```

connect the plug
```
snap connect codeverse-synth:home :home
```

Now you can play a wav file by doing:
```
sudo pulseaudio-example /var/snap/pulseaudio-example/common/<file.wav>
```
