A small proof-of-concept program to read the GPM103 based smart cards used in the Pokemon Snap Station.  A gemplus GCR410 smart card reader is required.

```ShellSession
[dogshoes@oxcart pokemon-snap-station-card-reader]# ./pokemon-snap-station-card-reader /dev/ttyUSB0
2015/04/10 18:16:51 Opened /dev/ttyUSB0.
2015/04/10 18:16:51 Setting sense type for GPM103 (0x7).
2015/04/10 18:16:52 Sense card.
2015/04/10 18:16:52 	Card is 5V.
2015/04/10 18:16:52 	Card is not powered.
2015/04/10 18:16:52 	Card is inserted.
2015/04/10 18:16:52 	Card protocol is T=0.
2015/04/10 18:16:52 	Detected card of type 0x7.
2015/04/10 18:16:52 Powering up the card.
2015/04/10 18:16:52 	Card ATR'd with 0x3B0000000000
2015/04/10 18:16:52 Serial number inquiry.
2015/04/10 18:16:52 	0x4601000559B8
2015/04/10 18:16:52 Counter inquiry.
2015/04/10 18:16:52 	Counter is 0x0000 credits.
```
