package main

const dashboardHTML = `<!DOCTYPE html>
<html lang="de" data-theme="light">
<head>
<style id="splash-vorab">
/* Bewusst als Erstes im Dokument: dieser Block ist wenige hundert Byte gross,
   deshalb kann der Browser das Startbild zeichnen, lange bevor das eigentliche
   Stylesheet durch ist. Vorher stand die Vorschau 33 KB weiter hinten — bis
   dahin sah man ein leeres Fenster. */
#splash{position:fixed;inset:0;z-index:9999;background:#0b0f14 center/cover no-repeat;
background-image:url(data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD//gARTGF2YzU4LjEzNC4xMDAA/9sAQwAUDg8SDw0UEhASFxUUGB4yIR4cHB49LC4kMklATEtHQEZFUFpzYlBVbVZFRmSIZW13e4GCgU5gjZeMfZZzfoF8/9sAQwEVFxceGh47ISE7fFNGU3x8fHx8fHx8fHx8fHx8fHx8fHx8fHx8fHx8fHx8fHx8fHx8fHx8fHx8fHx8fHx8fHx8/8AAEQgAGwAwAwEiAAIRAQMRAf/EAB8AAAEFAQEBAQEBAAAAAAAAAAABAgMEBQYHCAkKC//EALUQAAIBAwMCBAMFBQQEAAABfQECAwAEEQUSITFBBhNRYQcicRQygZGhCCNCscEVUtHwJDNicoIJChYXGBkaJSYnKCkqNDU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6g4SFhoeIiYqSk5SVlpeYmZqio6Slpqeoqaqys7S1tre4ubrCw8TFxsfIycrS09TV1tfY2drh4uPk5ebn6Onq8fLz9PX29/j5+v/EAB8BAAMBAQEBAQEBAQEAAAAAAAABAgMEBQYHCAkKC//EALURAAIBAgQEAwQHBQQEAAECdwABAgMRBAUhMQYSQVEHYXETIjKBCBRCkaGxwQkjM1LwFWJy0QoWJDThJfEXGBkaJicoKSo1Njc4OTpDREVGR0hJSlNUVVZXWFlaY2RlZmdoaWpzdHV2d3h5eoKDhIWGh4iJipKTlJWWl5iZmqKjpKWmp6ipqrKztLW2t7i5usLDxMXGx8jJytLT1NXW19jZ2uLj5OXm5+jp6vLz9PX29/j5+v/aAAwDAQACEQMRAD8AwJbtiflPFQ53n3NM4Aq1Y2puJgOi9zWhma2mQpbRGeQjIGaqtKt9K+9gvORmjUriOONYYWyF6n1rMaTCZ6GhsETXcqIRsl3MOMCqr3TSdcZqJhnkUzGDUFlhBntWzbv5GnllX5jxVJAN+MVbckqAenpVohmROW35OeaY5XYBnmr10BgcVmt96pZSH7wq4HJqInNKaUAYpDP/2Q==);
display:flex;align-items:flex-end;justify-content:center;transition:opacity .45s ease;}
</style>

<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Druckerfarm</title>
<link rel="icon" type="image/png" href="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADwAAABAEAYAAAAWjMiLAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAAABmJLR0QAAAAAAAD5Q7t/AAAACXBIWXMAAAEAAAABAAA610exAAAAB3RJTUUH6gkSCCcr9GF1vAAAFzdJREFUeNrtnXlYldXah+93TyCKKAQoapiiqGSSCTkrioLjMRwODpQDWWkeh0T4MlM7juhJ8pifAw6cJIfS9ISYWlpm4pCViTiQJIKy2RGTCGw2+32+P8Cw75yuRLNdXNz/rIu917vWs5/fWutdM4pUUE4FWmqpUWhsbUAtD5dagWs4tQLXcGoFruFUW+DbObdzbudAdmJ2YnYiGH8w/mD8AYxzjXONcyHPLc8tzw3Ufeo+dV/Vc2URZRFlEWB61/Su6V3IDs8Ozw4HY4IxwZgA2UnZSdlJULygeEHxgqrnCvwK/Ar84EflR+VHBdRANVANvCtdjzKPMg8wnTedN52HkmUly0qWVX0vJ+WknATTCtMK0wpI25q2NW0rZAVmBWYFglpHraPWqYpfoivRlejA9J7pPdN7YP7W/K3527vye6/svbL3wLTPtM+0D3KDcoNygyA3JDckNwRMvUy9TL3A1NnU2dQZTB1NHU0dITslOyU7BYqCi4KLgqEwsTCxMBGyD2QfyD4A2Q2zG2Y3BONm42bjZsg5nHM45zCo7qq76n7/AivV7UWvH7d+3PpxsHr36t2rd4NdsF2wXTAozZXmSnPQrNKs0qyCjo4dHTs6whzmMAcwxhnjjHEwNXRq6NRQ0Fg0Fo0F9M31zfXNwTLUMtQyFNxz3XPdc+H1iNcjXo+A7Z23d97eGVLUFDVFhX/t/9f+f+0Hj74efT36wtHMo5lHM2GGZYZlhgVe7fJql1e7QJ+0Pml90mDluZXnVp6DhIKEgoQCKIouii6KhjpedbzqeEFPx56OPR1hdsDsgNkBcFp3WndaB8t9lvss94GYlTErY1aCv7O/s78zzNsxb8e8HfBFiy9afNECprpOdZ3qCgcWHVh0YBEkm5PNyWawS7VLtUsF5WXlZeVlKD9afrT8KDy3/bntz22HHN8c3xxf2Ldk35J9S8Chm0M3h26gtFPaKe3Akm5Jt6RD/9P9T/c/DXOHzx0+dzjUD6sfVj/s3gXWVbdEmGJNsaZYSPNM80zzhOWxy2OXx0I7h3YO7RzgatDVoKtBsEy3TLdMB/lO+U75TjDq4KiDow7CRfNF80UzROyO2B2xGwauGrhq4Cq49u61d6+9C3979W+v/u1VWO2y2mW1C9yaf2v+rflw0fui90VvMIeYQ8whQAEFFMCt2bdm35oNKaEpoSmhkGXIMmQZYHGrxa0Wt4INyRuSNyTDyOKRxSOLoUePHj169ICbjjcdbzrC2vi18WvjoVBfqC/UQ/tr7a+1vwYp9VPqp9SHi74XfS/6wp6xe8buGQvxbvFu8W4w+4vZX8z+AgaEDwgfEA7/PPvPs/88C0XGImORsaqgOQ5xHOI4BOQZeUaeAa8ArwCvAFg6cenEpROhSIqkSGDphKUTlk4A5zHOY5zHwPZL2y9tvwQxhTGFMYUQMDlgcsBkCA4LDgt+mAIr9oq9Yg8GZ4OzwRk6t+ncpnMb8I/3j/ePh0ACCQQuvXbptUuvwa6Xdr206yXovK3zts7bQP+6/nX962BINCQaEsGgGBSDAvoN+g36DaDRarQaLRj6GfoZ+oEmWhOtiQZe4iVeAutE60TrxLsMWshCFoIuRheji4GsGVkzsmbAgeEHhh8YDp1OdDrR6QS87fG2x9seUG9hvYX1FlY9HrQ/aH/QfjDsNew17IUjXke8jniB0lpprbSGlZdXXl55GTJaZrTMaAkDDw08NPAQvDL6ldGvjAaLncXOYgfattq22rZQ6FfoV+gHu0fvHr17NBhGGUYZRoFugG6AbgBETI2YGjEVtHHaOG0caJ21zlpnMGgMGoMGDBcMFwwXQL9Mv0y/DHR2OjudHeg+0n2k+6i6at2HwAiCAAYMGIDxjGf8f4m2TtbJOuA85zkP0lt6S29QI9VINRI2/GXDXzb8Bbb13dZ3W1/I3Jq5NXMreH/u/bn35zB9w/QN0zfAojWL1ixaA2Uvlr1Y9iJY3rC8YXkDWMUqVkFpUmlSaRKox9Xj6nHQHdcd1x0HQ5ghzBAGljWWNZY1UPZE2RNlTwBP8iRPVr3b8o7nHc87Dg6NHRo7NAY5I2fkDKgL1AXqAigZXDK4ZDB4xnnGecbB6cjTkacj4Zj+mP6YHrrN7Daz20yQhbJQFoL91/Zf238NbY+3Pd72ONS5Xed2ndugLdGWaEvA0dXR1dEVNAM1AzUDIdM+0z7THmZ0mtFpRico7FDYobAD5Obn5ufmw5RWU1pNaQVdN3Xd1HUT0Ite9HqIAsthOSyHoTS8NLw0HA7GHYw7GAfpbdPbpreF1KjUqNQoeH/b+9ve3waD1gxaM2gNPPbBYx889gGUvVH2RtkbEL4lfEv4Fui/v//+/vth3MhxI8eNBOt163XrdXDv5t7NvRv4nPM553MOPuz/Yf8P+8MiWSSLBNrvbr+7/W7YP33/9P3TwbDQsNCwELr7dPfp7gMOix0WOyyGJX2W9FnSBybMmzBvwjzoWta1rGsZpLdMb5neEt7d/+7+d/fDMynPpDyTAt5HvY96HwX5Tr6T7yDiWMSxiGPg18+vn18/GOU0ymmUE0Q9H/V81PMQmxybHJsM4i3e4g3KJ8onyifQYF6DeQ3mQd2oulF1o0B9TH1MfQwub7m85fIWMO8x7zHvAcctjlsct8DGoxuPbjwKpx459cipRyAqJyonKgc0uzW7NbvB4GfwM/hVvwZrF1TweuXfv9qrvup51fOqJ2RcyriUcQlSe6X2Su0FSflJ+Un5kBmfGZ8ZD8PHDx8/fDxEto5sHdkaLMMswyzDIPlU8qnkUzB08NDBQwdDcK/gXsG9QDdRN1E3EZKbJTdLbgaN6jeq36g+jOo3qt+oflB+pPxI+RE463PW56wPfLn4y8VfLgaHmw43HW7CrB2zdszaASHmEHOIGTo91+m5Ts+Ba6hrqGsofJX+VfpX6fD5wc8Pfn4Qcs7nnM85D6FDQ4eGDoWoBlENohrArbduvXXrLchUMpVMBcJCwkLCQqC3R2+P3h5g8DJ4GbwgJSQlJCUEtPO087TzQGvUGrVGsHxm+czyGVzccnHLxS3wzaRvJn0zCb4+/vXxr4+DfrN+s34zuHzv8r3L96CMU8Yp42DSuUnnJp2DTlM6Tek0BdJc0lzSXOB6j+s9rveArt92/bbrt+Ay1GWoy9Dq1MjKXrTcI6VLS5eWLhXJ1+Xr8nUiueW55bnlIrkXci/kXhApblXcqrjVfz5n2WzZbNkskrc3b2/eXpHCrMKswiwRcz1zPXM9kaIdRTuKdogYZxlnGWeJ5FhzrDlWEUmVVEkVUYvUIrVIJLd3bu/c3iLph9MPpx8WyTuYdzDv4C/bWz6mfEz5GJEctxy3HDeRq+9cfefqOyI3xtwYc2OMiDXQGmgNrIpvfsn8kvklkbyIvIi8CJFSc6m51CwiTuIkTiLW3tbe1t4iBSsLVhasFCnIKMgoyBDJH58/Pn+8SHZCdkJ2gohppWmlaaVI3vd53+d9L5L/Wf5n+Z+J3DbdNt02iRQ3Lm5c3FikMKAwoDBAxJpvzbfmV9lRPL14evF0kXy3fLd8N5GyPmV9yvrcq0pVVFvgByU1PTU9NV1k5MaRG0duFAkuDC4MLhQJygrKCsoSCToSdCToiMjIVSNXjVwlktYurV1au6rn987ZO2fvHJHgVsGtgluJzJJZMktEbg+/Pfz28LsKYt3SuqV1RV5JeCXhlQSRwI6BHQM7ivTV99X31YsEtAxoGdBS5K9Zf836a5bI/sj9kfsjRcpGlI0oGyGyxGmJ0xInkYoWRGTa+mnrp60X+WHpD0t/WFqVj6pTdapOJHpI9JDoISJBXYK6BHURGV0wumB0gUjm3My5mXN/L+/+AQT+st6X9b6sJ1IxnhPx3+W/y3+XyBu8wRuILAxaGLQwSCR6YvTE6IkixmnGacZpIvnu+e757iK9XHu59nIVsQ+zD7MPE3Gc5DjJcZJI4pDEIYlDqvIpulR0qeiSSNcRXUd0HSHirfXWemtFlkcvj14eLbIicUXiikQRv0K/Qr9CES+Nl8ZLI3LBfMF8wSzSL7FfYr9EkTvtnHJNuaZcE9nUY1OPTT2q8kmemDwxeaKIR5pHmkdaVXw3BzcHNweRSwGXAi4F2E7g33+q8gxnOANaB62D1gGcmzo3dW4Kbeq2qdumLniv9V7rvRZ8m/s2920OLn93+bvL3+Ejw0eGjwxwYeOFjRc2wnyn+U7zncD9ivsV9ysQuzh2cexiMD9ifsT8CCinlFPKKVDmKHOUOeCpeqqeKowdMXbE2BHwbOKzic8mwlORT0U+FVk1A3dnXKrpr+mv6Q8eiz0WeyyGRgMbDWw0EDa9ventTW9DXp28Onl1IH5i/MT4iVDSoKRBSQPwedzncZ/HQROgCdAEABvZyMbf3cs/Uf1h0m+ExknjpHGCs7qzurM6uPn8zedvPg9WZ6uz1Rl8rvhc8bkCPk4+Tj5OsDVra9bWLNDt1O3U7QT9W/q39G+Bq8HV4GqAj10+dvnYBZJaJrVMagldMrpkdMkAbXNtc21zOLXj1I5TO2DYoWGHhh2CsillU8qmwI03b7x5400Yu3HsxrEbwXu+93zv+VDes7xneU9o8WKLF1u8CN3Cu4V3C4cYU4wpxgRvad/SvqWFXSd2ndh1Aob6DvUd6guaZppmmmaQMDhhcMJg2wn7k59/9xzf4z3eA0uIJcQSAoObDG4yuAkcDj0cejgUjqw9svbIWliXsS5jXQacHXB2wNkBcGLFiRUnVsCj0x6d9ug0+MTpE6dPnKBuSd2SuiVgbWptam0KW9pvab+lPZTeKr1VegtkqkyVqdBwW8NtDbfBzKyZWTOzoJmxmbGZEYoKiwqLCqH7zO4zu8+EulPrTq07FVS9qlf1oNRX6iv1IWxn2M6wneA5yXOS5yRYHrg8cHkgFJwsOFlwEia/M/mdye+AYbJhsmEyqA3VhmpD4ChHOWo7gX/3GqyP0EfoI6CxzmO8x3jwCGtS1qQM3La77XHbAzzN0zwN1htqA7UBHNpwOO5wHLQ76bPPZx9s+y5+RfwKaGpuqmuqA0uxxWgxQkTebN/ZvnAiPulM0hlILbn66dVPoX4TpzinOGh2RPlY+RiGNXrG/xl/8BjW5OUmL8PFmZemX5oOOy7szN2ZCz3Uno/0fARcfdx6u/UG9QP1mHoMWid6J3knwaTXwleEr4CYf8XExcTB4L8Nfnnwy+C3wX+9/3rYm74vc18mNP4fD28Pb9A9qhujG2M7gau92PCgmOPMceY4uHH6e9fvXcGxvW6QbhC4vug6wXUCANvYBuoKSx1LHcg8nTE+Yzxou2uf1T4LTa41sTaxAqsUN8UNQBmvjIf8v+cV5xVDzrc5+Tn54BboNsFtAmTHZh/KPgTl71tPWU9B65OtbrS6ATzLq7wK6Xuvn71+FtQfVINqgKaPNO3QtAPkTv9x+Y/LQZJIJx2aXGpS0KQAikOLZxbPhJvjbs6+ORtcilzqudSDhmucFWcFTPHZLbJbwO0lZf8o+wc09fVe570O9PEVU7I1XuAqDg44OABgwdwFcwFUi2oBYBObAKzZ1mwApbPSGYD61AdQZ6uzoaLxBMAbbwCln9IPQBmtjAaQj+QjAM0WzRYAxjIWQD2tngagN70BNHYaOwA60AFAnpfnAZSxylgAWtEKKhp7AL7jOwBNgiYBQG7LbQDJl3wAtZ/aD6Dn0Z5HAaJDo0MBdE10TX4/797BhgK/M+KdEQATgiYEAUgvqZhjbUELAKWj0hEAe+wBsGIFQEX9WVjdz+/1+/sMrWarGWBQs0HNAD4Y+MFAAP06/brfz7t3sFkvGpQzyhkA7QDtAABxERcApa5S13ZWPThqmVoGoBmgGQBAQxrazhpbbtl5nMcBsMPOhlbUcGxYg4klFoAUUgAqlyGrjzPOAOpwdTiANdoafT/JKEuUJQBaF21FS7JAWQBAKaU29NIDYssmulSpcFwb2gD3LbCqqApAh9gOsQCBbwa+CcBtbldEqHxH/pId3ZXuAAWPFjwKsMu4ywhQ6FnoCaBcVi7bzksPig0F1nTRdAFQqBg9yDk5dz/pWLdbtwP4F/oXAiwzLDMAVG5I+HUuchEgwz7DHuBQw0MNAQrqF9S/274/JzZ8ByvPKs8C0IUuwK/WtF+kHvWgaphyv1h3W3cDyAfyAQB69Lbzzm+FLQWerEwGHlzg34r77QP8obGhwJoRmhEAdKZiIsPWAtdIbFmD2yntAPDFF6ihNcjW2FLgHCUHKjaoArUCPxRsOdFROedMYxoDtQI/FGxZg6cr04FagR8qtpzoaKO0AShfVL4IwLrWuhaAz/js5xErB6J3Qk1lsawM5Sl5CkB1Vp0BCKMaRzvuYjazAawZ1gyA8uLyYgDlS+XL6iQjBjEAWHtaewJgxGgrH9tU4JaXW14GeGHFCysAyh3KHQDoTncASZd0AGWwUrH1ZQlLAJS3lbcBiCQSQA1Xw6FqeY4SSgCoQ517MqQSRy9HL4AxyWOSAXI35W4CUEyKqTrpqH3VvgDtY9rHAGiSNcm287INlwsfGnea+j/3FNRvRE08AF4r7F3URIFruYtagWs4tQLXcGoFruHUClzDseE4uOLuDkhKTkoGKNeWa6Fqh8W9Iv+WfwN4bvXcCvDEyidW3o89Jc4lzgBfZH2RBVCSWpIKoAxVqnMe9yd73P3d/QH8+vr1BVA+VD78XRz7M2w4Dj716alPAQacHnAa4JZ6SwVQjinHqpOOaEULEJ4RngHwv9/87zf3Y8+12GuxAH0X9V0EcL399fYAilWx3o89wTOCZwDsObDnAIB+pf6+Ct6DYcMafGc+onxv+V6A8qTyJIBqn+UJIADAOt06/YEMmsIUgHK3cjeA8oTyhAey50nrkwBkkfXwvPhr2HJXZeVJBXpSMWerrWg/lD5Kn+ok89MaxZ3Ne/fLnQ32XekKVByTexB7Kk9KkEbaw3Lhr1Pbyarh1Apcw6kVuIZjS4GXsxxA5st8oMYtEvx0mC6PPNtZYUOB7a/YXwGwe9/ufQA+53MAdDbt+lWfyoIpRvnZwn69rHpZAIqX4mU742wocNt1bdcBvOb/mj+AU7JTMoBYpOKcsLZ6o3LNYc3hB7FHe157Hqi4/fKeMqzcUfKpfAqAO+4AnWM6xwBE+UX5Aeg66Do8fG/+EjasK3ZP2z0NMI1pADiHOYcBzBk2ZxhA9t7svQCKRqkohr+0Z2sMYwBS/p3yb4D1O9fvBKrOJv2qIRWnG3987cfXAIpeKHoBgElMAmAzm38Wv7LgyQJZAEAjGgEMujDoAkDF1Q7gVer1Rzi0dmcmy3Y3Of3EeBkvIrJv6L6hIiKPjXhshEjFzVMiIpRRJiKiKIpyd0jlZV9sY9vP4lc3fJRHRUTIIuu/5aPoFb2ICK1pLSKiFa2IiEzoOKGjiIgxyBhkaxf+J38kgf8fnzb7tJmIyOMHHj8gcpcQJkwiIopG0YjcJbQVq4gIJZTcV1hKqYiIgsLd6f4kbGX+dYbVGSYiEpkWmSYiUuha6GprX/0yf2CB7/BV6FehIiJd3Lq4idx7jVQMikHkv9TEO2Hl96xj3b2k1yC3Qa6IyJs337wpImJ2Njvb2je/zp9o013qldQrABv8N/gDFLgXVPwvgxd5EUB5RXkF4Fu7b+0ATq4+uRqA56m4VOXOu/zOy2m9rAdw2OqwFWBws8HNAJxed6q4e3cHOwCUncpOgH5R/aIAQvJC8gA0szSzbO2Te+COwLYuab8dK+atmCdy7zXddbXrahGRy+sur7O17b89f64R5z3RyK2RG8BTIU+FACgXlAvAf4yvxVd8AVwGugwEsB9mPwyAF3jB1r/ht+NP1ETfK+Yz5jMApWNLK+65qhyf/ge55AIoi5RFAPVy6+UCaCZpJtn6N/x21ECBa7mb2sWGGk6twDWc/wO4eOiap8OMcwAAACV0RVh0ZGF0ZTpjcmVhdGUAMjAyNi0wOS0xOFQwODozOTo0MyswMDowMN7Nf2IAAAAldEVYdGRhdGU6bW9kaWZ5ADIwMjYtMDktMThUMDg6Mzk6NDMrMDA6MDCvkMfeAAAAPnRFWHRzdmc6Y29tbWVudAAgQ3JlYXRlZCB3aXRoIElua3NjYXBlIChodHRwOi8vd3d3Lmlua3NjYXBlLm9yZy8pIP5FR7MAAAAASUVORK5CYII=">
<style>
:root, html[data-theme="dark"] {
  --bg:#000000; --s1:#000000; --s2:#0b0b0b; --s3:#161616;
  --btn-bg:#161616; --btn-text:#ffffff;
  --seg-bg:#161616;
  --menu-bg:#0b0b0b; --menu-text:#ffffff;
  --border-menu:#3a3a3a;
  --border:#333333; --border2:#4a4a4a;
  --accent:#00d5ff; --adim:rgba(0,213,255,0.18);
  --green:#00c750; --gdim:rgba(0,199,80,0.18);
  --orange:#ff9d1f; --odim:rgba(255,157,31,0.18);
  --red:#ff5470; --rdim:rgba(255,84,112,0.18);
  --gold:#ffd700;
  /* Schrift bewusst nur zwei Stufen, beide klar lesbar auf Schwarz.
     Vorher war --muted #4a5968 — auf dunklem Grund kaum zu entziffern. */
  --text:#ffffff; --muted:#c8c8c8; --muted2:#a0a0a0;
}
html[data-theme="light"] {
  --bg:#ffffff; --s1:#ffffff; --s2:#f4f4f4; --s3:#e9e9e9;
  --btn-bg:#e9e9e9; --btn-text:#000000;
  --seg-bg:#e9e9e9;
  --menu-bg:#f4f4f4; --menu-text:#000000;
  --border-menu:#c2c2c2;
  --border:#cfcfcf; --border2:#a8a8a8;
  --accent:#0069c2; --adim:rgba(0,105,194,0.12);
  --green:#00c750; --gdim:rgba(0,199,80,0.12);
  --orange:#a85800; --odim:rgba(168,88,0,0.12);
  --red:#c1123a; --rdim:rgba(193,18,58,0.1);
  --gold:#8a6d00;
  --text:#000000; --muted:#3d3d3d; --muted2:#5a5a5a;
}
*{box-sizing:border-box;margin:0;padding:0;}
body{background:var(--bg);color:var(--text);font-family:Arial,Helvetica,system-ui,sans-serif;min-height:100vh;overflow-x:hidden;}
html[data-theme="dark"] body::after{content:'';position:fixed;inset:0;background:repeating-linear-gradient(0deg,transparent,transparent 2px,rgba(255,255,255,0.02) 2px,rgba(255,255,255,0.02) 4px);pointer-events:none;z-index:9999;}
header{display:flex;align-items:center;background:var(--s1);border-bottom:1px solid var(--border);position:sticky;top:0;z-index:100;height:52px;}
.logo{display:flex;align-items:center;gap:10px;padding:0 18px;height:100%;border-right:1px solid var(--border);min-width:175px;}
.logo-mark{width:28px;height:28px;background:var(--accent);border-radius:4px;display:grid;place-items:center;font-size:14px;flex-shrink:0;}
.logo h1{font-size:13px;font-weight:800;letter-spacing:-0.3px;line-height:1.2;}
.logo h1 span{color:var(--accent);}
.header-tabs{display:flex;height:100%;flex:1;overflow:hidden;}
.tab{padding:0 15px;height:100%;display:flex;align-items:center;gap:7px;font-size:12px;font-weight:600;color:var(--muted);cursor:pointer;border-right:1px solid var(--border);transition:all 0.15s;white-space:nowrap;}
.tab:hover{color:var(--text);background:var(--s2);}
.tab.active{color:var(--accent);background:var(--adim);border-bottom:2px solid var(--accent);}
.g2dot{width:7px;height:7px;border-radius:50%;background:var(--muted);flex-shrink:0;transition:all 0.3s;}
#g2-proc.warn{color:var(--red);font-weight:800;}
.g2dot.online{background:var(--green);box-shadow:0 0 6px rgba(57,255,126,0.6);animation:pls 2s infinite;}
.g2dot.offline{background:var(--red);}
@keyframes pls{0%,100%{opacity:1}50%{opacity:0.6}}
.header-right{display:flex;align-items:center;gap:6px;padding:0 12px;margin-left:auto;border-left:1px solid var(--border);height:100%;}
.grid-sizer{display:flex;align-items:center;gap:7px;font-size:11px;color:var(--muted);font-family:ui-monospace,Consolas,monospace;}
.grid-sizer input[type=range]{width:160px;accent-color:var(--accent);cursor:pointer;}
.hbtn{height:30px;padding:0 12px;border-radius:4px;font-size:12px;font-weight:700;cursor:pointer;border:1px solid var(--border2);background:var(--s2);color:var(--text);transition:all 0.15s;white-space:nowrap;}
.hbtn:hover{border-color:var(--accent);color:var(--accent);}
.hbtn.primary{background:var(--accent);color:#000;border-color:var(--accent);}
.hbtn.primary:hover{background:#00c8e0;}
.hbtn.green{background:var(--green);color:#000;border-color:var(--green);}
.hbtn.red{border-color:var(--red);color:var(--red);}
.hbtn.red:hover{background:var(--red);color:#fff;}
/* Lang switcher */
.lang-sw{display:flex;gap:2px;align-items:center;}
.lang-btn{height:26px;padding:0 8px;border-radius:3px;font-size:11px;font-weight:700;cursor:pointer;border:1px solid var(--border2);background:transparent;color:var(--muted);transition:all 0.15s;font-family:ui-monospace,Consolas,monospace;letter-spacing:0.5px;}
.lang-btn:hover{color:var(--text);border-color:var(--border2);}
.lang-btn.active{background:var(--adim);border-color:var(--accent);color:var(--accent);}
main{padding:10px 10px 40px;}
.view{display:none;}.view.active{display:block;}
.top-bar{display:flex;gap:7px;align-items:center;flex-wrap:wrap;margin:unset;margin-bottom:0;padding:8px 10px;background:var(--seg-bg);}
.stat-pill{background:var(--s1);border:1px solid var(--border);border-radius:5px;padding:6px 13px;display:flex;align-items:center;gap:8px;cursor:pointer;transition:all 0.15s;user-select:none;}
.stat-pill:hover{border-color:var(--border2);}
.jp-result{background:var(--s2);border:1px solid var(--border);border-radius:6px;padding:8px 10px;font-size:13px;line-height:1.6;}
.ai-out{margin-top:6px;background:var(--s3);border:1px solid var(--border);border-radius:6px;padding:8px 10px;font-size:12px;line-height:1.55;white-space:pre-wrap;}
.ai-ta{flex:1;min-width:0;width:100%;resize:vertical;font-family:inherit;font-size:12px;line-height:1.4;}
.ai-conn-wrap{flex:1;min-width:0;display:flex;flex-direction:column;gap:8px;}
.ai-conn-list{display:flex;flex-direction:column;gap:6px;}
.ai-conn-empty{color:var(--muted);font-size:12px;padding:4px 2px;}
.ai-conn-row{display:flex;align-items:center;gap:10px;flex-wrap:wrap;background:var(--s3);border:1px solid var(--border);border-radius:6px;padding:6px 10px;}
.ai-conn-row.active{border-color:var(--green);box-shadow:inset 0 0 0 1px var(--green);}
.ai-conn-pick{display:flex;align-items:center;gap:6px;cursor:pointer;}
.ai-conn-name{font-weight:600;}
.ai-conn-meta{color:var(--muted);font-size:12px;flex:1;min-width:120px;}
.ai-conn-btns{display:flex;gap:6px;}
.ai-conn-add{display:flex;gap:6px;flex-wrap:wrap;align-items:center;}
.ai-conn-add .repo-input{flex:1;min-width:140px;}
.ai-conn-edit{flex:1;min-width:0;display:flex;flex-direction:column;gap:6px;background:var(--s3);border:1px solid var(--border);border-radius:6px;padding:10px;}
.ai-edit-f{display:flex;align-items:center;gap:8px;}
.ai-edit-f>label{width:130px;flex:none;font-size:12px;color:var(--muted);}
.ai-edit-f .repo-input,.ai-edit-f .set-select{flex:1;min-width:0;}
.pc-btn.danger{color:var(--red);}
/* Auth: user chip + admin access card + role hiding */
.user-chip{display:flex;align-items:center;gap:8px;padding:0 6px 0 10px;margin-left:6px;border-left:1px solid var(--border);}
.user-name{font-weight:600;font-size:13px;}
.user-role{font-size:11px;color:var(--muted);text-transform:uppercase;letter-spacing:.04em;}
.admin-only{display:none;}
body.role-admin .admin-only{display:block;}
body:not(.role-admin) #tab-import{display:none;}
.set-sub{font-size:13px;margin:16px 0 6px;color:var(--fg);}
.acc-list{display:flex;flex-direction:column;gap:6px;margin-bottom:8px;}
.acc-row{display:flex;align-items:center;gap:8px;flex-wrap:wrap;background:var(--s3);border:1px solid var(--border);border-radius:6px;padding:5px 10px;}
.acc-name{font-weight:600;min-width:120px;flex:1;}
.acc-role{max-width:140px;}
.acc-role-tag{font-size:12px;color:var(--muted);text-transform:uppercase;}
.acc-empty{color:var(--muted);font-size:12px;}
.acc-add{display:flex;gap:6px;flex-wrap:wrap;align-items:center;margin-bottom:6px;}
.acc-add .repo-input{flex:1;min-width:120px;}
.token-reveal{background:var(--s2);border:1px solid var(--accent);border-radius:6px;padding:8px 10px;margin-top:8px;font-size:12px;word-break:break-all;}
.token-reveal code{font-family:ui-monospace,Consolas,monospace;color:var(--accent);}
.jp-line{}
.mf-row{display:flex;align-items:center;gap:8px;margin-bottom:4px;}
.mf-model{font-family:ui-monospace,Consolas,monospace;font-size:11px;font-weight:700;min-width:64px;color:var(--accent);}
.mf-input{flex:1;min-width:0;}
.sync-diff-hint{background:rgba(255,140,0,0.10);border:1px solid rgba(255,140,0,0.35);border-radius:6px;padding:8px 10px;font-size:12px;}
.mf-diff-title{font-weight:700;margin-bottom:4px;}
.mf-diff-row{font-family:ui-monospace,Consolas,monospace;font-size:11px;line-height:1.5;}
.stat-pill.active-all{border-color:var(--text);background:rgba(220,232,240,0.07);}
.stat-pill.active-model{border-color:var(--mc);background:color-mix(in srgb,var(--mc) 12%,transparent);}
.stat-pill.active-online{border-color:var(--accent);color:var(--accent);background:var(--adim);font-weight:700;}
.stat-pill.alarm-live{border-color:var(--red);background:rgba(255,51,85,0.14);animation:alarmPulse 1.1s ease-in-out infinite;}
@keyframes alarmPulse{0%,100%{box-shadow:0 0 0 0 rgba(255,51,85,0.0);}50%{box-shadow:0 0 0 3px rgba(255,51,85,0.28);}}
.sl{font-size:11px;font-family:ui-monospace,Consolas,monospace;color:var(--muted);text-transform:uppercase;letter-spacing:0.8px;}
.sv{font-size:20px;font-weight:900;font-family:ui-monospace,Consolas,monospace;line-height:1;}
.sv.ct{color:var(--text);}.sv.cm{color:var(--mc,var(--accent));}
.search-input{flex:1;min-width:130px;max-width:220px;height:34px;padding:0 10px;background:var(--s1);border:1px solid var(--border);border-radius:5px;color:var(--text);font-family:ui-monospace,Consolas,monospace;font-size:13px;}
.search-input:focus{outline:none;border-color:var(--accent);}
.stream-bar{display:flex;gap:6px;align-items:center;flex-wrap:wrap;padding:8px 10px;background:var(--seg-bg);border:unset;margin:unset;border-radius:0;border-bottom:1px solid var(--border);}
.stream-bar-label{font-size:10px;font-family:ui-monospace,Consolas,monospace;color:var(--muted);text-transform:uppercase;letter-spacing:1px;margin-right:4px;}
.sbtn{height:28px;padding:0 12px;border-radius:4px;font-size:11px;font-weight:800;cursor:pointer;border:none;transition:all 0.15s;white-space:nowrap;display:flex;align-items:center;gap:5px;}
.sbtn:hover{filter:brightness(1.15);transform:translateY(-1px);}
.sbtn.s-play{background:var(--green);color:#000;}
.sbtn.s-stop{background:var(--red);color:#fff;}
.sbtn.s-fav{background:var(--gold);color:#000;}
.sbtn.s-snap{background:var(--accent);color:#000;}
.res-btn.res-active{background:var(--accent);color:#000;outline:2px solid var(--accent);}
.stream-divider{width:1px;height:20px;background:var(--border2);margin:0 2px;}
.sub-nav{display:flex;align-items:center;gap:0;background:var(--seg-bg);border-bottom:1px solid var(--border);}
.filter-label{font-size:10px;font-family:ui-monospace,Consolas,monospace;color:var(--muted);text-transform:uppercase;letter-spacing:1px;margin-right:2px;}
.pill-host{display:contents;}
.sv-green{color:var(--green);}
.sv-red{color:var(--red);}
.sv-gold{color:var(--gold);}
.sv-amber{color:#e0a020;}
.stream-count{font-family:ui-monospace,Consolas,monospace;font-size:10px;color:var(--muted);margin-left:auto;}
#printerGrid{display:grid;grid-template-columns:repeat(6,1fr);gap:6px;align-items:start;}
.tile{background:var(--s1);border:1px solid var(--border);border-radius:6px;overflow:visible;display:flex;flex-direction:column;min-width:0;transition:border-color 0.2s,box-shadow 0.2s;animation:fadeIn 0.25s ease both;}
@keyframes fadeIn{from{opacity:0;transform:scale(0.96)}to{opacity:1;transform:scale(1)}}
.tile:hover{border-color:var(--border2);box-shadow:0 2px 14px rgba(0,0,0,0.5);}
.tile.fav{border-color:rgba(255,215,0,0.35);}
.tile.streaming{}
.tile.has-error{border-color:rgba(255,51,85,0.6);background:rgba(255,51,85,0.06);}
.tile.has-error .tile-header{background:rgba(255,51,85,0.12);}

.tile-live-badge.dim{background:var(--muted2);animation:none;color:var(--muted);}
.tile-live-badge.offline{background:var(--muted2);animation:none;color:var(--muted);}
.mdot{width:6px;height:6px;border-radius:50%;flex-shrink:0;margin-top:1px;}
.tile-name{font-size:13px;font-weight:700;flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.tile-live-badge{background:var(--red);color:#fff;font-family:ui-monospace,Consolas,monospace;font-size:9px;font-weight:700;padding:2px 5px;border-radius:3px;letter-spacing:0.5px;flex-shrink:0;}
.fav-star.on{opacity:1;filter:drop-shadow(0 0 5px rgba(255,215,0,0.9));}
.tile-cam{position:relative;aspect-ratio:16/9;background:#020406;overflow:hidden;width:100%;}
.cam-pill{background:color-mix(in srgb,var(--muted) 18%,transparent);color:var(--muted);}
.bt-table{width:100%;border-collapse:collapse;font-size:12px;margin-top:6px;}
.bt-table th{text-align:left;color:var(--muted);font-weight:600;padding:4px 6px;border-bottom:1px solid var(--border);}
.bt-table td{padding:3px 6px;border-bottom:1px solid var(--border);}
.bt-in{width:100%;box-sizing:border-box;background:var(--s2);color:var(--fg);border:1px solid var(--border);border-radius:4px;padding:3px 6px;font-size:12px;}
.cam-tile .cam-spacer .df-balken-fuell{width:0;}
.cam-center>.cam-idle,.cam-center>.cam-wait,.cam-center>.cam-repair{display:flex;flex-direction:column;align-items:center;justify-content:center;height:100%;}
.tile-fs{position:absolute;right:6px;bottom:6px;z-index:5;width:26px;height:26px;border:none;border-radius:5px;background:rgba(0,0,0,0.55);color:#fff;font-size:14px;line-height:1;cursor:pointer;opacity:0;transition:opacity 0.15s;display:flex;align-items:center;justify-content:center;}
.tile-cam:hover .tile-fs{opacity:1;}
.tile-fs:hover{background:rgba(0,0,0,0.8);}
.scan-cell-fixed:hover .fsv-cell-fs,.fsv-cell:hover .fsv-cell-fs{opacity:1;}
.tile-video{position:absolute;top:0;left:0;width:100%;height:100%;object-fit:cover;}
.tile-cam iframe{position:absolute;top:50%;left:50%;transform:translate(-50%,-50%);width:calc(100% + 20px);height:calc(100% + 40px);border:none;scrollbar-width:none;}
.cam-idle{position:absolute;inset:0;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:6px;background:#020406;}
.cam-idle .ci{font-size:26px;opacity:0.25;}
.cam-idle .cip{font-family:ui-monospace,Consolas,monospace;font-size:10px;color:var(--muted);text-align:center;line-height:1.9;}
.cam-wait,.cam-connect{position:absolute;inset:0;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:8px;background:#020406;z-index:2;}
.cam-connect{background:rgba(2,4,6,0.65);}
.cam-idle.ready .ci{opacity:0.6;color:var(--green);}
.cam-spinner{width:22px;height:22px;border-radius:50%;border:2.5px solid rgba(255,255,255,0.2);border-top-color:rgba(255,255,255,0.85);animation:snap-dreh 0.9s linear infinite;}
.cam-repair{position:absolute;inset:0;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:6px;background:transparent;color:var(--muted);}
.cam-repair .ci{font-size:30px;line-height:1;}
.cam-repair .cip{font-size:15px;font-weight:800;letter-spacing:1px;text-transform:uppercase;}
.cam-kein-bild{background:#020406;}
.cam-kein-bild .cam-spinner{display:none;}
.cam-offline{background:#020406;color:var(--muted);}
.cam-private{background:#020406;color:var(--muted);}
.fsv-hinweis.private{background:transparent;color:var(--muted);}
.ql-ping{cursor:pointer;}
.ql-ping:hover{filter:brightness(1.25);}
.ql-actions{display:flex;gap:6px;justify-content:flex-end;align-items:center;flex-wrap:nowrap;}
.cam-badge{flex-shrink:0;margin-left:6px;padding:1px 7px;border-radius:10px;font-size:9px;font-weight:800;letter-spacing:0.3px;background:var(--s3);color:var(--muted);text-transform:uppercase;}
.cam-dot{background:var(--muted)!important;}
.cam-form{display:flex;gap:8px;flex-wrap:wrap;align-items:center;}
.cam-in{flex:1;min-width:160px;height:30px;padding:0 10px;background:var(--s1);border:1px solid var(--border);border-radius:5px;color:var(--text);font-size:13px;}
.set-select{height:30px;padding:0 10px;background:var(--s1);border:1px solid var(--border);border-radius:5px;color:var(--text);font-size:13px;}
.cam-row{display:flex;gap:12px;align-items:center;padding:5px 0;border-bottom:1px solid var(--border);}
.cam-row-name{font-weight:600;min-width:140px;}
.cam-row-src{flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:var(--muted);}
.xiaomi-box{border:1px solid var(--border);border-radius:6px;padding:8px 10px;background:var(--s1);}
.xiaomi-box summary{cursor:pointer;font-weight:700;font-size:13px;}
.xm-result{white-space:pre-wrap;word-break:break-word;max-height:180px;overflow:auto;margin-top:8px;color:var(--muted);font-family:ui-monospace,Consolas,monospace;}
.xm-captcha{max-height:60px;border-radius:4px;background:#fff;padding:2px;margin-bottom:6px;}
.fsv-hinweis{position:absolute;inset:0;display:flex;align-items:center;justify-content:center;z-index:2;text-align:center;font-weight:800;letter-spacing:1px;text-transform:uppercase;font-size:clamp(14px,2.2vw,26px);}
.fsv-hinweis.repair{background:transparent;color:var(--muted);}
.fsv-hinweis.offline{background:#020406;color:var(--muted);}
#setup-icon.arbeitet{color:var(--gold);text-shadow:0 0 8px var(--gold);animation:zahnrad-puls 1.1s ease-in-out infinite;}
@keyframes zahnrad-puls{0%,100%{opacity:0.45;}50%{opacity:1;}}
.snap-container{position:absolute;inset:0;}
.snap-container img{width:100%;height:100%;object-fit:cover;display:block;}
.snap-age{position:absolute;bottom:4px;left:5px;font-family:ui-monospace,Consolas,monospace;font-size:8px;color:rgba(255,255,255,0.4);background:rgba(0,0,0,0.5);padding:1px 4px;border-radius:2px;}
@keyframes lbpls{0%,100%{opacity:1}50%{opacity:0.6}}
.cam-loading{position:absolute;inset:0;display:flex;align-items:center;justify-content:center;background:#020406;}
.spinner{width:20px;height:20px;border:2px solid var(--border2);border-top-color:var(--accent);border-radius:50%;animation:spin 0.8s linear infinite;}
@keyframes spin{to{transform:rotate(360deg)}}
.tile-acts{display:flex;gap:3px;flex-shrink:0;}
.tact{width:24px;height:24px;background:var(--s3);border:1px solid var(--border);border-radius:3px;color:var(--muted);font-size:12px;cursor:pointer;display:flex;align-items:center;justify-content:center;transition:all 0.15s;}
.tact:hover{border-color:var(--accent);color:var(--accent);}
.tact.t-play:hover{border-color:var(--green);color:var(--green);}
.tact.t-stop:hover,.tact.t-del:hover{border-color:var(--red);color:var(--red);}
.tact.active-play{border-color:var(--green);color:var(--green);background:var(--gdim);}
.sbtn.res-active{border-color:var(--accent);color:var(--accent);background:var(--adim);font-weight:700;}
.list-item.fav{border-color:rgba(255,215,0,0.25);}
.mono{font-family:ui-monospace,Consolas,monospace;font-size:11px;}
.lact{height:24px;padding:0 9px;background:var(--s2);border:1px solid var(--border);border-radius:4px;color:var(--muted);font-size:10px;font-family:ui-monospace,Consolas,monospace;cursor:pointer;transition:all 0.15s;white-space:nowrap;}
.lact:hover{border-color:var(--accent);color:var(--accent);}
.lact.ld:hover{border-color:var(--red);color:var(--red);}
.lstar.on{opacity:1;}
.ip{max-width:1200px;margin:0 auto;}
.sp{max-width:860px;margin:0 auto;}
.pc,.stcard{background:var(--s1);border:1px solid var(--border);border-radius:8px;padding:22px;margin-bottom:14px;}
.pc h2,.stcard h2,.stcard h3{font-size:14px;font-weight:800;margin-bottom:4px;display:flex;align-items:center;gap:10px;}
.pc>p,.stcard>p{font-size:13px;color:var(--muted);margin-bottom:14px;line-height:1.6;}
.csvfmt{background:var(--s2);border:1px solid var(--border2);border-radius:6px;padding:11px 14px;font-family:ui-monospace,Consolas,monospace;font-size:12px;color:var(--accent);margin-bottom:14px;line-height:2;}
.csvfmt .cm{color:var(--muted);}
textarea{width:100%;background:var(--s2);border:1px solid var(--border2);border-radius:6px;color:var(--text);font-family:ui-monospace,Consolas,monospace;font-size:12px;padding:11px;resize:vertical;min-height:140px;line-height:1.8;transition:border-color 0.2s;}
textarea:focus{outline:none;border-color:var(--accent);}
.ia{display:flex;gap:10px;margin-top:10px;align-items:center;}
.ir{font-size:12px;font-family:ui-monospace,Consolas,monospace;padding:5px 11px;border-radius:4px;display:none;}
.ir.ok{background:var(--gdim);color:var(--green);border:1px solid rgba(57,255,126,0.2);}
.ir.er{background:var(--rdim);color:var(--red);border:1px solid rgba(255,51,85,0.2);}
.frow{display:grid;grid-template-columns:110px 1fr 1fr 1fr auto;gap:9px;align-items:end;}
.fg{display:flex;flex-direction:column;gap:5px;}
label{font-size:10px;font-family:ui-monospace,Consolas,monospace;color:var(--muted);text-transform:uppercase;letter-spacing:0.8px;}
input,select{background:var(--s2);border:1px solid var(--border2);border-radius:5px;color:var(--text);padding:8px 10px;font-family:ui-monospace,Consolas,monospace;font-size:12px;transition:border-color 0.2s;width:100%;}
input::placeholder{color:var(--muted2);}
input:focus,select:focus{outline:none;border-color:var(--accent);}
select option{background:var(--s2);}
.dz{border:2px dashed var(--border2);border-radius:8px;padding:22px;text-align:center;cursor:pointer;transition:all 0.2s;margin-bottom:10px;}
.dz:hover,.dz.drag{border-color:var(--accent);background:var(--adim);}
.dz p{font-size:12px;color:var(--muted);}
.dz strong{color:var(--accent);}
#fi{display:none;}
.status-row{display:flex;align-items:center;gap:14px;}
.sind{width:14px;height:14px;border-radius:50%;flex-shrink:0;}
.sind.checking{background:var(--muted);animation:pls 1s infinite;}
.sind.online{background:var(--green);box-shadow:0 0 8px rgba(57,255,126,0.5);animation:pls 2s infinite;}
.sind.offline{background:var(--red);}
.sinfo{flex:1;}
.sinfo h3{font-size:14px;font-weight:800;margin-bottom:2px;display:block;}
.sinfo p{font-size:12px;color:var(--muted);margin:0;}
.snum{width:22px;height:22px;background:var(--accent);color:#000;border-radius:50%;display:grid;place-items:center;font-size:11px;font-weight:900;flex-shrink:0;}
.cb{background:var(--bg);border:1px solid var(--border2);border-radius:6px;padding:12px 14px;font-family:ui-monospace,Consolas,monospace;font-size:12px;color:var(--accent);line-height:1.9;position:relative;margin-top:8px;}
.yout{background:var(--bg);border:1px solid var(--border2);border-radius:6px;padding:14px;font-family:ui-monospace,Consolas,monospace;font-size:11px;color:var(--text);line-height:2;max-height:300px;overflow-y:auto;white-space:pre;margin-top:8px;}
.yk{color:var(--accent);}.yv{color:#a8ff78;}.ycm{color:var(--muted);}
.toasts{position:fixed;bottom:20px;right:20px;z-index:500;display:flex;flex-direction:column;gap:6px;}
.toast{background:var(--s2);border:1px solid var(--border2);border-radius:6px;padding:9px 13px;font-size:12px;min-width:240px;box-shadow:0 4px 20px rgba(0,0,0,0.5);animation:tIn 0.2s ease,tOut 0.25s ease 2.75s forwards;display:flex;align-items:center;gap:8px;font-family:ui-monospace,Consolas,monospace;}
.toast.ok{border-color:rgba(57,255,126,0.3);}.toast.er{border-color:rgba(255,51,85,0.3);}
@keyframes tIn{from{transform:translateX(14px);opacity:0}to{transform:translateX(0);opacity:1}}
@keyframes tOut{to{opacity:0}}

/* Printer status bar */
.ts-fill.running{background:var(--green);}
.ts-fill.paused{background:var(--orange);}
.ts-fill.failed{background:var(--red);}
.ts-fill.finish{background:var(--green);}
/* State pill in header */
.tile-state{padding:1px 6px;border-radius:3px;font-size:9px;font-weight:800;letter-spacing:0.5px;font-family:ui-monospace,Consolas,monospace;flex-shrink:0;}
.tile-state.running{background:var(--gdim);color:var(--green);}
.tile-state.pause{background:var(--odim);color:var(--orange);}
.tile-state.failed{background:var(--rdim);color:var(--red);}
.tile-state.finish{background:var(--adim);color:var(--accent);}
.ts-offline{color:var(--muted);font-size:9px;padding:3px 0;}

/* Error display */

/* Dropdown info */
.tile-name-wrap{position:relative;width:100%;overflow:hidden;}
.tile-name{font-size:13px;font-weight:700;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;cursor:pointer;display:flex;align-items:center;gap:4px;}
.tile-name:hover{color:var(--accent);}
.tile-name-arrow{font-size:11px;opacity:0.7;flex-shrink:0;color:var(--accent);}
.tile-name-arrow .arr{display:inline-block;transition:transform 0.15s;}
.tile-name-arrow.open .arr{display:inline-block;transform:rotate(180deg);}
.tile-dropdown{display:none;position:absolute;top:100%;left:-7px;right:-7px;z-index:300;background:var(--s3);border:1px solid var(--border2);border-top:none;border-radius:0 0 6px 6px;padding:8px 10px;font-size:11px;font-family:ui-monospace,Consolas,monospace;box-shadow:0 6px 20px rgba(0,0,0,0.7);}
.tile-dropdown.open{display:block;}
/* Das Menue haengt am body statt in der Kachel: .tile-name-wrap und
   .tile-header-row1 haben overflow:hidden und haben es vorher komplett
   weggeschnitten — deshalb war "mehr" ohne jede Wirkung. */
#dd-pop{display:none;position:fixed;z-index:9400;min-width:230px;max-width:560px;background:var(--s3);border:1px solid var(--border2);border-radius:6px;padding:9px 11px;font-size:11px;font-family:ui-monospace,Consolas,monospace;box-shadow:0 10px 34px rgba(0,0,0,0.8);}
#dd-pop.open{display:block;}
#dd-backdrop{display:none;position:fixed;inset:0;z-index:9390;background:rgba(0,0,0,0.5);}
#dd-backdrop.open{display:block;}
#dd-pop.dd-modal{position:fixed;left:50%;top:50%;transform:translate(-50%,-50%);width:min(900px,calc(100% - 32px));max-width:none;max-height:calc(100% - 48px);overflow:auto;}
#dd-pop .dd-title{font-size:12px;font-weight:700;color:var(--text);margin-bottom:6px;font-family:system-ui,sans-serif;display:flex;align-items:center;justify-content:space-between;gap:10px;}
.ams-unit{margin:5px 0;}
.ams-head{color:var(--muted);font-size:10px;margin-bottom:3px;}
.ams-trays{display:flex;gap:5px;flex-wrap:wrap;}
.ams-tray{display:flex;align-items:center;gap:4px;background:var(--s2);border:1px solid var(--border);border-radius:4px;padding:2px 5px;font-size:10px;height:20px;box-sizing:border-box;}
.ams-tray.leer{opacity:0.5;}
.ams-tray .ams-dot{flex:0 0 11px;}
.ams-tray-typ{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.ams-tray.active{border:2px solid #22c55e;box-shadow:0 0 0 2px rgba(34,197,94,.30);font-weight:600;padding:1px 4px;}
.ams-tray.active .ams-active-mark{color:#22c55e;font-size:9px;line-height:1;margin-right:1px;}
.ams-tray.geladen{border-color:var(--accent);border-style:dashed;}
.dd-sel{background:var(--s2);color:var(--fg);border:1px solid var(--border);border-radius:4px;font-size:11px;height:24px;padding:0 4px;max-width:120px;}
.ams-dot{width:11px;height:11px;border-radius:3px;border:1px solid rgba(255,255,255,0.35);flex:none;}
.ams-dot.empty{background:repeating-linear-gradient(45deg,#2a2f36,#2a2f36 3px,#1a1e24 3px,#1a1e24 6px);}
.pc-row{flex-wrap:wrap;display:flex;gap:5px;margin-top:8px;}
.tc-block,.fl-block{margin-top:6px;}
.tc-zeile{display:flex;align-items:center;gap:4px;margin-bottom:4px;font-size:11px;}
.tc-lbl{width:52px;flex:none;color:var(--muted);}
.tc-ist{width:38px;text-align:right;font-family:ui-monospace,Consolas,monospace;color:var(--muted);}
.tc-feld{width:46px;background:var(--s3);border:1px solid var(--border2);border-radius:4px;color:var(--text);padding:2px 4px;font-size:11px;text-align:center;}
.tc-btn{background:var(--s2);border:1px solid var(--border2);border-radius:4px;color:var(--text);padding:2px 7px;font-size:11px;cursor:pointer;}
.tc-btn:hover{border-color:var(--accent);}
.tc-set{font-size:10px;}
.fl-titel{font-size:10px;text-transform:uppercase;letter-spacing:0.6px;color:var(--muted);margin-bottom:4px;}
.fl-farbe{width:40px;height:22px;border:1px solid var(--border2);border-radius:4px;background:none;cursor:pointer;padding:0;}
.fl-schmal{width:56px;}
.pc-btn{flex:0 0 auto;border:1px solid var(--border2);background:var(--s2);color:var(--text);border-radius:4px;padding:5px 6px;font-size:11px;cursor:pointer;font-family:system-ui,sans-serif;}
.pc-btn:hover:not(:disabled){border-color:var(--accent);}
.pc-btn:disabled{opacity:0.35;cursor:default;}
.pc-btn.danger{color:var(--red);border-color:rgba(255,51,85,0.45);}
.pc-btn.sm{padding:3px 6px;font-size:10px;text-decoration:none;display:inline-block;}
.tl-switch{display:inline-flex;align-items:center;gap:5px;cursor:pointer;font-size:11px;}
.media-ov{position:fixed;inset:0;background:rgba(0,0,0,0.55);display:flex;align-items:center;justify-content:center;z-index:9700;}
.media-box{background:var(--s1);border:1px solid var(--border2);border-radius:8px;width:min(680px,92vw);max-height:82vh;display:flex;flex-direction:column;box-shadow:0 10px 40px rgba(0,0,0,0.4);}
.media-head{display:flex;align-items:center;justify-content:space-between;padding:12px 14px;border-bottom:1px solid var(--border2);font-weight:600;}
.media-x{background:none;border:none;color:var(--text);font-size:16px;cursor:pointer;line-height:1;}
.media-info{padding:8px 14px;color:var(--muted);font-size:12px;border-bottom:1px solid var(--border2);}
.media-list{overflow:auto;padding:6px 10px;flex:1 1 auto;}
.media-row{display:flex;align-items:center;gap:8px;padding:5px 4px;border-bottom:1px solid var(--border);}
.media-name{flex:1 1 auto;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:12px;}
.media-size{flex:0 0 auto;color:var(--muted);font-size:11px;min-width:64px;text-align:right;}
.media-foot{display:flex;gap:8px;justify-content:space-between;padding:10px 14px;border-top:1px solid var(--border2);}
.tile.vp-minimal .df-fuss{display:none!important;}
.tt-row{display:flex;justify-content:space-between;gap:12px;padding:2px 0;color:var(--text);}
.tt-label{color:var(--muted);white-space:nowrap;}
.tt-val{color:var(--text);text-align:right;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.tt-err{color:var(--red);padding:3px 0;margin-top:2px;}
.tt-err::before{content:"⚠ ";}
.tt-sep{border-top:1px solid var(--border);margin:5px 0;}
.tt-empty{color:var(--muted);font-size:10px;}
.dd-row{display:flex;justify-content:space-between;gap:10px;padding:2px 0;}.dd-lbl{color:var(--muted);white-space:nowrap;font-size:10px;}.dd-val{color:var(--text);text-align:right;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}.dd-val-failed{color:var(--red);font-weight:700;}
.dd-preview{margin:6px 0;text-align:center;}
.dd-preview-img{max-width:100%;max-height:160px;border-radius:6px;border:1px solid var(--border);background:#0b0f14;}.dd-sep{border-top:1px solid var(--border);margin:5px 0;}.dd-err{color:var(--red);padding:2px 0;font-size:10px;}.dd-err::before{content:"⚠ ";}.dd-none{color:var(--muted);font-size:10px;}
/* Edit row */
.list-item.editing{border-color:var(--accent);background:var(--s2);}
.list-item.editing input{background:var(--s3);border:1px solid var(--border2);border-radius:4px;color:var(--text);padding:3px 7px;font-family:ui-monospace,Consolas,monospace;font-size:11px;width:100%;}
.list-item.editing input:focus{outline:none;border-color:var(--accent);}
.lact.le:hover{border-color:var(--accent);color:var(--accent);}
.lact.ls:hover{border-color:var(--green);color:var(--green);}

/* SD card list */
.sd-list{margin-top:10px;}
.sd-printer-row{display:flex;align-items:center;gap:10px;padding:8px 12px;background:var(--s2);border:1px solid var(--border);border-radius:5px;margin-bottom:4px;cursor:pointer;transition:border-color 0.15s;user-select:none;}
.sd-printer-row:hover{border-color:var(--accent);}
.sd-printer-row.open{border-color:var(--accent);border-bottom-left-radius:0;border-bottom-right-radius:0;}
.sd-printer-name{font-weight:700;font-size:13px;flex:1;}
.sd-file-count{font-family:ui-monospace,Consolas,monospace;font-size:11px;color:var(--accent);white-space:nowrap;}
.sd-size{font-family:ui-monospace,Consolas,monospace;font-size:11px;color:var(--muted);white-space:nowrap;}
.sd-err{font-family:ui-monospace,Consolas,monospace;font-size:10px;color:var(--red);}
.fs-hint{margin:6px 0 10px;padding:8px 11px;border:1px solid var(--border2);border-left:3px solid var(--accent);border-radius:4px;background:var(--s2);color:var(--muted);font-size:12px;line-height:1.45;}
.ftpc-row{display:grid;grid-template-columns:20px minmax(120px,1fr) 130px 2fr;align-items:center;gap:8px;padding:4px 8px;border-bottom:1px solid var(--border);font-size:12px;}
.ftpc-row .ftpc-name{font-weight:600;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.ftpc-row .ftpc-port{color:var(--muted);font-family:ui-monospace,Consolas,monospace;font-size:11px;}
.ftpc-row .ftpc-reason{color:var(--muted);}
.ftpc-row.ok .ok{color:var(--green);} .ftpc-row.warn .warn{color:var(--amber,#e0a000);} .ftpc-row.err .err{color:var(--red);}
.ftpc-row.err{background:rgba(255,51,85,0.07);} .ftpc-row.warn{background:rgba(224,160,0,0.07);}
.sd-arrow{font-size:10px;color:var(--muted);transition:transform 0.15s;flex-shrink:0;}
.sd-printer-row.open .sd-arrow{transform:rotate(180deg);}
.sd-files{display:none;background:var(--s3);border:1px solid var(--accent);border-top:none;border-radius:0 0 5px 5px;margin-bottom:4px;max-height:260px;overflow-y:auto;overflow-x:hidden;}
.sd-files.open{display:block;}
.sd-file-tools{display:flex;align-items:center;gap:8px;padding:4px 6px;border-bottom:1px solid var(--border);margin-bottom:3px;font-size:11px;}
.sd-check{display:flex;align-items:center;gap:4px;color:var(--muted);cursor:pointer;}
.sd-sel-info{color:var(--muted);}
.sd-fcb{flex:none;margin-right:2px;}
.sd-file-row{display:grid;grid-template-columns:18px minmax(0,1fr) auto auto auto;align-items:center;gap:8px;padding:5px 12px;border-bottom:1px solid var(--border);font-family:ui-monospace,Consolas,monospace;font-size:11px;}
.sd-file-row:last-child{border-bottom:none;}
.sd-file-row.red{grid-template-columns:1fr;color:var(--red);}
.sd-kind-seg{display:flex;gap:4px;padding:7px 12px 3px;}
.sd-kind-btn{border:1px solid var(--border2);background:var(--s2);color:var(--muted);border-radius:4px;padding:3px 10px;font-size:11px;cursor:pointer;}
.sd-kind-btn:hover{border-color:var(--accent);}
.sd-kind-btn.on{background:var(--adim);border-color:var(--accent);color:var(--accent);font-weight:700;}
.sep-note{grid-column:auto;font-size:9px;color:var(--muted);white-space:nowrap;}
.sd-file-name{min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:var(--text);}
.sd-file-size{color:var(--muted);white-space:nowrap;}
.sd-dl-btn{height:20px;padding:0 7px;background:var(--s2);border:1px solid var(--border);border-radius:3px;color:var(--muted);font-size:9px;cursor:pointer;transition:all 0.15s;white-space:nowrap;}
.sd-dl-btn:hover{border-color:var(--accent);color:var(--accent);}
.sd-loading{text-align:center;padding:20px;color:var(--muted);font-size:12px;font-family:ui-monospace,Consolas,monospace;}
.sd-summary-bar{display:flex;gap:16px;padding:8px 12px;background:var(--s1);border:1px solid var(--border);border-radius:5px;margin-bottom:10px;font-family:ui-monospace,Consolas,monospace;font-size:11px;}
.sd-summary-item{display:flex;flex-direction:column;gap:2px;}
.sd-summary-val{font-size:16px;font-weight:900;color:var(--text);}
.sd-summary-lbl{color:var(--muted);font-size:9px;text-transform:uppercase;letter-spacing:0.8px;}

/* Sync upload drop zone */
.sync-filter-bar{display:flex;align-items:center;gap:8px;flex-wrap:wrap;margin-bottom:10px;}
.sync-model-pill{padding:3px 10px;border-radius:12px;font-size:11px;font-weight:700;cursor:pointer;border:1px solid var(--border2);background:var(--s2);color:var(--muted);transition:all 0.15s;font-family:ui-monospace,Consolas,monospace;}
.sync-model-pill.active{border-color:var(--accent);background:var(--adim);color:var(--accent);}
.sd-search{padding:0 10px;}
.sd-search:not(:empty){padding:4px 10px 8px 30px;border-bottom:1px solid var(--border);}
.fs-hit{display:flex;gap:10px;align-items:center;padding:4px 6px;font-size:12px;border-radius:4px;}
.fs-hit:hover{background:var(--s2);}
.fs-hit input[type=checkbox]{width:auto;accent-color:var(--accent);flex-shrink:0;}
.fs-hit .fs-file{font-family:ui-monospace,Consolas,monospace;color:var(--text);flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.fs-hit .fs-size{color:var(--muted);font-family:ui-monospace,Consolas,monospace;white-space:nowrap;}
.sd-hitcount{font-size:10px;font-family:ui-monospace,Consolas,monospace;color:var(--accent);margin-left:6px;white-space:nowrap;}
.sd-hitcount.none{color:var(--muted);}
.sd-hitcount.fail{color:var(--red);}
.ql-row,.ql-head{display:grid;grid-template-columns:16px 64px 84px minmax(180px,1fr) 96px 150px 120px 360px;gap:10px;align-items:center;padding:5px 0;border-bottom:1px solid var(--border);}
.ql-head{font-size:10px;text-transform:uppercase;letter-spacing:0.6px;color:var(--muted);font-family:ui-monospace,Consolas,monospace;}
.ql-head span.sortbar{cursor:pointer;user-select:none;white-space:nowrap;}
.ql-head span.sortbar:hover{color:var(--text);}
.ql-head span.aktiv{color:var(--accent);}
.ql-wrap{overflow-x:auto;}
.ql-wrap .ql-row,.ql-wrap .ql-head{min-width:1120px;}
.ql-row>*{min-width:0;}
.ql-code{overflow:hidden;text-overflow:ellipsis;letter-spacing:1px;}
.ql-row .ql-name{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.ql-row .ql-serial{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.ql-row.editing{background:var(--s2);padding:8px 6px;border-radius:5px;}
.ql-row input{height:28px;font-size:12px;}
/* Das Logo ist schwarz auf weisser Flaeche. Im dunklen Schema wuerde daraus ein
   weisser Kasten, deshalb dort invertiert. */
.app-logo{height:26px;width:auto;display:block;}
html[data-theme="dark"] .app-logo{filter:invert(1) hue-rotate(180deg);}
/* Punkt 1: Kacheln heben sich im hellen Schema vom Untergrund ab */
html[data-theme="light"] .tile{background:var(--seg-bg);}
.sbtn.s-play{color:#fff;text-shadow:0 0 1px #000;}
.hbtn.primary{color:#fff;}
/* Punkt 20: Herkunftsangabe neben dem Logo, bewusst unuebersetzt */
.app-suffix{font-size:11px;font-weight:600;letter-spacing:1px;color:var(--menu-text);opacity:0.75;white-space:nowrap;}
.comp-sub{margin-top:12px;padding-left:14px;border-left:2px solid var(--border2);}
.comp-sub .status-row{padding:0;}
.cu-row{display:flex;gap:10px;align-items:center;font-size:12px;padding:3px 0;}
.cu-row .cu-new{color:var(--accent);font-weight:700;}
.cu-row .cu-name{margin-right:8px;}
.set-grid{display:flex;flex-direction:column;gap:12px;margin-top:12px;}
.set-row{display:flex;align-items:center;gap:14px;flex-wrap:wrap;}
.set-row > label{font-size:13px;color:var(--text);min-width:190px;}
.seg{display:inline-flex;border:1px solid var(--border2);border-radius:5px;overflow:hidden;}
.seg-btn{background:var(--seg-bg);color:var(--muted);border:none;padding:6px 16px;font-size:12px;cursor:pointer;font-family:inherit;}
.seg-btn + .seg-btn{border-left:1px solid var(--border2);}
.seg-btn.active{background:var(--accent);color:var(--bg);font-weight:700;}
.scan-row{display:grid;grid-template-columns:22px 70px minmax(0,1fr) 150px 160px 70px 200px;gap:8px;align-items:center;padding:6px 8px;border:1px solid var(--border);border-radius:5px;margin-bottom:5px;background:var(--s1);font-size:12px;}
.scan-row>span{overflow:hidden;text-overflow:ellipsis;}
.scan-row.bekannt{opacity:0.85;}
.scan-haken{text-align:center;color:var(--green);}
.scan-row.umgezogen .scan-haken{color:var(--gold);}
.scan-abschnitt{margin-bottom:14px;}
.scan-abschnitt-titel{font-size:9px;text-transform:uppercase;letter-spacing:1px;color:var(--muted);margin:0 0 5px 2px;}
.scan-row.umgezogen{border-color:var(--gold);background:rgba(255,215,0,0.06);}
.scan-bar{display:flex;align-items:center;gap:10px;flex-wrap:wrap;padding:8px 10px;background:var(--seg-bg);border:1px solid var(--border);border-radius:5px;margin-bottom:8px;}
.scan-code{width:100%;background:var(--s3);border:1px solid var(--border2);border-radius:4px;color:var(--text);padding:3px 6px;font-family:ui-monospace,Consolas,monospace;font-size:11px;}
.scan-check{cursor:pointer;}
.scan-row.known{opacity:0.5;}
.scan-model{font-weight:700;font-family:ui-monospace,Consolas,monospace;}
.scan-mono{font-family:ui-monospace,Consolas,monospace;color:var(--muted);overflow:hidden;text-overflow:ellipsis;}
.scan-tag{font-size:10px;padding:1px 6px;border-radius:9px;border:1px solid var(--border2);color:var(--muted);text-align:center;}
.scan-tag.free{color:var(--green);border-color:rgba(57,255,126,0.4);}
.scan-tag.occupied{color:var(--orange);border-color:rgba(255,165,0,0.4);}
.drop-zone{border:2px dashed var(--border2);border-radius:8px;padding:20px;text-align:center;color:var(--muted);font-size:12px;cursor:pointer;transition:all 0.2s;background:var(--s1);margin-bottom:12px;}
.drop-zone.drag-over{border-color:var(--accent);background:var(--adim);color:var(--accent);}
.drop-zone.has-files{border-color:var(--green);background:var(--gdim);}
.drop-file-list{margin-top:8px;text-align:left;}
.drop-file-item{display:flex;align-items:center;justify-content:space-between;padding:3px 6px;background:var(--s3);border-radius:3px;margin-bottom:2px;font-family:ui-monospace,Consolas,monospace;font-size:10px;}
.drop-file-name{color:var(--text);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;flex:1;}
.drop-file-rm{color:var(--muted);cursor:pointer;padding:0 4px;flex-shrink:0;}
.drop-file-rm:hover{color:var(--red);}
.sync-printer-list{display:flex;flex-direction:column;}
.sync-printer-item{display:flex;align-items:center;gap:8px;padding:6px 10px;border-bottom:1px solid var(--border);font-size:12px;}
.sync-printer-item:last-child{border-bottom:none;}
.sync-printer-item input[type=checkbox]{width:auto;accent-color:var(--accent);flex-shrink:0;}
.sync-printer-label{flex:1;}
.sync-upload-btn{padding:6px 16px;border-radius:4px;border:1px solid var(--green);background:var(--gdim);color:var(--green);font-size:12px;font-weight:700;cursor:pointer;transition:all 0.15s;}
.sync-upload-btn:hover{background:var(--green);color:#000;}
.sync-upload-btn:disabled{opacity:0.4;cursor:not-allowed;}
.sync-progress-item{display:flex;align-items:center;gap:8px;padding:3px 0;font-size:11px;font-family:ui-monospace,Consolas,monospace;}
.sync-pi-name{width:80px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:var(--muted);}
.sync-pi-bar{flex:1;height:4px;background:var(--border2);border-radius:2px;overflow:hidden;}
.sync-pi-fill{height:100%;background:var(--green);transition:width 0.3s;}
.sync-pi-status{font-size:10px;color:var(--muted);white-space:nowrap;}

.tab-dropdown-wrap{position:relative;}
.tab-dropdown-menu{position:absolute;top:100%;left:0;z-index:500;background:var(--s3);border:1px solid var(--border2);border-radius:0 0 6px 6px;min-width:130px;box-shadow:0 4px 16px rgba(0,0,0,0.5);}
.tab-dd-item{padding:9px 16px;font-size:13px;cursor:pointer;color:var(--text);display:flex;align-items:center;gap:8px;transition:background 0.1s;}
.tab-dd-item:hover{background:var(--s2);color:var(--accent);}
.tab-dd-item.active{color:var(--accent);}

.vp-item{padding:8px 16px;font-size:12px;cursor:pointer;color:var(--text);}
.vp-item:hover{background:var(--s2);color:var(--accent);}
.vp-item.active{color:var(--accent);}
.tile.vp-minimal .df-kopf,.tile.vp-minimal .df-pillen,.tile.vp-minimal .df-fuss{display:none!important;}

.color-grid{display:flex;flex-direction:column;gap:8px;}
.color-row{display:flex;align-items:center;gap:10px;padding:4px 0;}
.color-label{width:160px;font-size:12px;color:var(--text);flex-shrink:0;}
.color-row input[type=color]{width:36px;height:30px;border:1px solid var(--border2);border-radius:4px;background:none;cursor:pointer;padding:2px;flex-shrink:0;}
.color-row input[type=text]{width:90px;font-family:ui-monospace,Consolas,monospace;font-size:12px;background:var(--s2);color:var(--text);border:1px solid var(--border2);border-radius:4px;padding:4px 8px;}
.color-preview{width:24px;height:24px;border-radius:4px;border:1px solid var(--border2);flex-shrink:0;}
::-webkit-scrollbar{width:5px}::-webkit-scrollbar-track{background:var(--bg)}::-webkit-scrollbar-thumb{background:var(--border2);border-radius:3px}

.fs-mode header,.fs-mode .header-tabs-row,.fs-mode #sub-nav,.fs-mode #stream-bar,.fs-mode #top-bar{display:none!important;}
.fs-mode #view-fsv,.fs-mode #view-fsc{position:fixed;top:0;left:0;right:0;bottom:0;z-index:7000;overflow-y:auto;background:#000;}
.fs-esc{position:fixed;top:8px;right:12px;z-index:8000;background:rgba(0,0,0,0.75);color:#fff;border:1px solid rgba(255,255,255,0.3);border-radius:4px;padding:5px 14px;font-size:12px;cursor:pointer;}
.fsv-cell{position:relative;aspect-ratio:16/9;background:#000;overflow:hidden;}
.fsv-cell video{width:100%;height:100%;object-fit:cover;display:block;}
.fsv-label{position:absolute;bottom:0;left:0;right:0;padding:2px 6px;font-size:10px;color:rgba(255,255,255,0.85);background:linear-gradient(transparent,rgba(0,0,0,0.7));}
.fsv-grid{display:grid;grid-template-columns:repeat(var(--cols,9),1fr);gap:2px;background:#111;padding:2px;}
.sbtn.res-active{border-color:var(--accent);color:var(--accent);background:var(--adim);font-weight:700;}

.list-item.editing{grid-template-columns:20px 1fr 1fr 1fr 1fr 1fr auto;cursor:default;}
/* Die Detailzeile haengt direkt unter ihrer Zeile — kein schwebendes Menue,
   das an Fensterraendern verrutscht. Standardmaessig ist sie nicht da. */
.list-detail{display:none;padding:8px 12px 12px 44px;background:var(--s2);border-bottom:1px solid var(--border);}
.list-detail.auf{display:block;}
/* Gleiche Hoehe fuer alle vier Kaesten: align-items:stretch statt start, und
   die Kaesten selbst als Spalte, damit ihr Inhalt den Platz ausfuellt. */
.dk-raster{display:grid;grid-template-columns:repeat(auto-fit,minmax(230px,1fr));gap:10px;align-items:stretch;}
.dk{display:flex;flex-direction:column;}
.dk .dk-inhalt{flex:1;}
.dk{background:var(--s1);border:1px solid var(--border);border-radius:6px;overflow:hidden;}
.dk-titel{font-size:9px;text-transform:uppercase;letter-spacing:1px;color:var(--muted);padding:5px 9px;background:var(--s3);border-bottom:1px solid var(--border);}
.dk-inhalt{padding:7px 9px 9px;}
.dk.dk-video .dk-inhalt{padding:0;}
.dk-zeile{display:flex;justify-content:space-between;gap:10px;padding:2px 0;font-size:11px;align-items:center;}
.dk-bez{color:var(--muted);flex:none;}
.dk-wert{text-align:right;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.dk-wert.mono{font-family:ui-monospace,Consolas,monospace;font-size:10px;}
.dk-leer{font-size:11px;color:var(--muted);}
.dk-balken{display:inline-block;width:52px;height:5px;background:var(--border2);border-radius:3px;overflow:hidden;vertical-align:middle;}
.dk-balken-fuell{display:block;height:100%;width:var(--w,0%);background:var(--accent);}
.dk-balken-fuell.running{background:var(--green);}
.dk-balken-fuell.pause{background:var(--gold);}
.dk-balken-fuell.failed{background:var(--red);}
.dk-ams{margin-bottom:8px;}
.dk-ams:last-child{margin-bottom:0;}
.dk-ams-kopf{font-size:10px;font-weight:700;color:var(--accent);margin-bottom:2px;}
.dk-fehler{margin-top:6px;font-size:10px;color:var(--red);line-height:1.4;}
.dk .pc-row{margin-top:8px;}
.dk-eingabe{background:var(--s3);border:1px solid var(--border2);border-radius:4px;color:var(--text);padding:3px 6px;font-size:11px;width:150px;text-align:right;}
.dk-eingabe.mono{font-family:ui-monospace,Consolas,monospace;font-size:10px;}
.dk-eingabe:focus{outline:none;border-color:var(--accent);}
.dk-videotest{padding:7px 9px;display:flex;flex-direction:column;gap:4px;align-items:flex-start;}
.kt-gut{color:var(--green);}
.kt-schlecht{color:var(--gold);}
.ld-cam{position:relative;width:100%;aspect-ratio:16/9;background:#000;border-radius:5px;overflow:hidden;}
.ld-cam img{width:100%;height:100%;object-fit:cover;display:block;}
.ld-cam img.hidden{display:none;}
.ld-cam-warten{position:absolute;inset:0;display:flex;align-items:center;justify-content:center;}
.ld-cam-warten.hidden{display:none;}
.ld-cam-fehler{font-size:10px;color:var(--muted);text-align:center;padding:0 8px;line-height:1.4;}

.list-detail .dd-row{display:flex;justify-content:space-between;gap:14px;padding:2px 0;font-size:12px;max-width:520px;}
.list-detail .dd-lbl{color:var(--muted);}
.list-detail .dd-val{font-family:ui-monospace,Consolas,monospace;}
.dd-fein{font-size:9px;color:var(--muted);opacity:0.8;}
.list-detail .dd-sep{height:1px;background:var(--border);margin:6px 0;max-width:520px;}
.list-detail .ld-akt{display:flex;gap:6px;margin-top:8px;flex-wrap:wrap;}
.list-item .ll-pfeil{display:inline-block;width:10px;color:var(--muted);transition:transform 0.15s;}
.list-item.auf .ll-pfeil{transform:rotate(90deg);color:var(--accent);}
.list-item.auf{background:var(--s2);}
.list-tools{display:flex;align-items:center;gap:8px;padding:6px 12px;}

/* ── Aus dem Markup ueberfuehrte Stile ───────────────────────────────────── */
.note{font-size:12px;color:var(--muted);}
.note-sm{font-size:11px;color:var(--muted);}
.note-xs{font-size:10px;color:var(--muted);}
.muted{color:var(--muted);}
.red{color:var(--red);}
.lh{line-height:1.6;}
.mono{font-family:ui-monospace,Consolas,monospace;}
.mono-xs{font-size:10px;color:var(--muted);font-family:ui-monospace,Consolas,monospace;}
.mono-note{font-family:ui-monospace,Consolas,monospace;font-size:12px;color:var(--muted);}
.mono-value{font-family:ui-monospace,Consolas,monospace;font-size:13px;}
.breakable{word-break:break-all;}
.fs12{font-size:12px;}
.fs12-text{font-size:12px;color:var(--text);}
.hidden{display:none!important;}
/* Ladeanzeige */
@keyframes df-spin{to{transform:rotate(360deg);}}
.spinner{display:inline-block;width:14px;height:14px;border:2px solid var(--border2);border-top-color:var(--accent);border-radius:50%;animation:df-spin 0.7s linear infinite;vertical-align:middle;}
.spinner.big{width:26px;height:26px;border-width:3px;}
.spinner.spinner-sm{width:11px;height:11px;border-width:2px;}
/* Startbild: liegt über allem, blendet nach zwei Sekunden weich aus. */
#splash{position:fixed;inset:0;z-index:9999;background:#000 center/cover no-repeat;display:flex;align-items:flex-end;justify-content:center;transition:opacity 0.45s ease;}
#splash.weg{opacity:0;pointer-events:none;}
/* Das scharfe Bild liegt als eigene Ebene darueber und wird eingeblendet,
   sobald es geladen ist. Darunter steht ab dem ersten Byte die 337 Byte
   grosse Vorschau — dadurch ist sofort ein Bild da statt einer schwarzen
   Flaeche, und es wird scharf statt zu erscheinen. */
#splash-scharf{position:absolute;inset:0;background:center/cover no-repeat;opacity:0;transition:opacity 0.35s ease;}
#splash-scharf.da{opacity:1;}
/* Das Über-Fenster zeigt dasselbe Bild wie der Start, aber mittig und mit dem
   Copyright aus den Einstellungen. */
#ueber{position:fixed;inset:0;z-index:9998;background:#000 center/cover no-repeat;display:flex;align-items:center;justify-content:center;cursor:pointer;}
#ueber.hidden{display:none;}
#ueber-inhalt{display:flex;align-items:center;justify-content:center;}
#ueber-text{text-align:center;padding:26px 40px;background:rgba(0,0,0,0.55);border-radius:8px;backdrop-filter:blur(4px);}
.ueber-logo{height:46px;filter:invert(1) grayscale(2);}
.ueber-suffix{margin:2px 0 14px;color:#fff;font-size:12px;letter-spacing:3px;text-transform:uppercase;font-family:ui-monospace,Consolas,monospace;}
.ueber-copy{margin:0;color:#fff;font-size:13px;}
.ueber-version{margin:6px 0 0;color:rgba(255,255,255,0.6);font-size:11px;font-family:ui-monospace,Consolas,monospace;}
.ueber-hinweis{margin:16px 0 0;color:rgba(255,255,255,0.45);font-size:11px;}
#splash{align-items:center !important;}
#splash-mitte{text-align:center;padding:26px 40px;background:rgba(0,0,0,0.5);border-radius:10px;backdrop-filter:blur(4px);}
.splash-logo{height:44px;filter:invert(1) grayscale(2);}
.splash-suffix{margin:3px 0 16px;color:#fff;font-size:12px;letter-spacing:3px;text-transform:uppercase;font-family:ui-monospace,Consolas,monospace;}
.splash-copy{margin:16px 0 0;color:rgba(255,255,255,0.75);font-size:12px;}
.oss-hinweis{margin:10px 0 0;color:var(--muted);font-size:10px;line-height:1.6;}
.oss-link{color:var(--accent);text-decoration:underline;}
.splash-ver{margin:4px 0 0;color:rgba(255,255,255,0.55);font-size:11px;font-family:ui-monospace,Consolas,monospace;}
.lang-picker{position:fixed;inset:0;z-index:10001;background:#000 center/cover no-repeat;background-image:url(/splash.jpg?v=__BUILD_ID__);display:flex;align-items:center;justify-content:center;}
.lp-box{text-align:center;padding:34px 44px;background:rgba(0,0,0,0.6);border-radius:12px;backdrop-filter:blur(6px);}
.lp-logo{height:52px;width:auto;margin-bottom:16px;filter:invert(1) hue-rotate(180deg);}
.lp-title{color:#fff;font-size:20px;font-weight:800;line-height:1.35;margin-bottom:22px;}
.lp-sub{color:rgba(255,255,255,0.6);font-size:14px;font-weight:600;}
.lp-btns{display:flex;gap:14px;justify-content:center;flex-wrap:wrap;}
.lp-btn{min-width:150px;padding:14px 22px;font-size:15px;font-weight:800;cursor:pointer;border:1px solid rgba(255,255,255,0.25);border-radius:8px;background:rgba(255,255,255,0.08);color:#fff;transition:all 0.15s;}
.lp-btn:hover{background:var(--accent);border-color:var(--accent);color:#001018;}
.praesenz-banner{background:var(--gold);color:#111;font-size:12px;font-weight:800;text-align:center;padding:6px 12px;letter-spacing:0.3px;}
.jobs-badge{display:inline-flex;align-items:center;justify-content:center;min-width:16px;height:16px;padding:0 4px;margin-left:6px;border-radius:8px;background:var(--accent);color:#001018;font-size:10px;font-weight:800;}
.jobs-panel{position:fixed;top:0;right:0;width:360px;max-width:92vw;height:100vh;z-index:9000;background:var(--s1);border-left:1px solid var(--border);box-shadow:-8px 0 24px rgba(0,0,0,0.35);display:flex;flex-direction:column;}
.jobs-page{display:flex;gap:18px;align-items:flex-start;flex-wrap:wrap;padding:4px 2px;}
.jobs-col{background:var(--s1);border:1px solid var(--border);border-radius:8px;padding:14px 16px;}
.jobs-col-run{flex:0 0 320px;max-width:340px;}
.jobs-col-plan{flex:1;min-width:340px;}
.jobs-col h2{margin:0 0 10px;font-size:15px;}
.plan-head{display:flex;justify-content:space-between;align-items:center;gap:10px;flex-wrap:wrap;}
.plan-head h2{margin:0;}
.plan-head-btns{display:flex;gap:6px;}
.plan-add{display:flex;gap:6px;margin:10px 0;flex-wrap:wrap;}
.plan-add .repo-input{flex:1;min-width:160px;}
.plan-empty{color:var(--muted);font-size:13px;padding:10px 2px;}
.plan-matrix-wrap{overflow-x:auto;}
.plan-table{border-collapse:collapse;width:100%;font-size:13px;}
.plan-table th,.plan-table td{border:1px solid var(--border);padding:4px 6px;text-align:center;}
.plan-table thead th{background:var(--s3);position:sticky;top:0;white-space:nowrap;}
.plan-table .plan-fig-h{text-align:left;}
.plan-table .plan-fig{text-align:left;min-width:160px;}
.plan-name{width:100%;min-width:140px;}
.plan-qty{width:56px;text-align:center;padding:3px 4px;border:1px solid var(--border);border-radius:4px;background:var(--s2);color:var(--fg);font-family:ui-monospace,Consolas,monospace;}
.plan-sum{font-weight:700;font-family:ui-monospace,Consolas,monospace;background:var(--s3);}
.plan-table tfoot td{background:var(--s3);font-weight:700;}
.jobs-head{display:flex;align-items:center;justify-content:space-between;padding:14px 16px;border-bottom:1px solid var(--border);}
.jobs-title{font-size:14px;font-weight:800;}
.jobs-list{flex:1;overflow-y:auto;padding:12px 14px;display:flex;flex-direction:column;gap:10px;}
.jobs-empty{color:var(--muted);font-size:12px;text-align:center;padding:24px 0;}
.job-item{background:var(--s2);border:1px solid var(--border);border-radius:7px;padding:10px 12px;}
.job-item.done{opacity:0.6;}
.job-top{display:flex;align-items:center;justify-content:space-between;gap:8px;margin-bottom:7px;}
.job-label{font-size:12px;font-weight:700;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.job-cancel{flex:none;border:1px solid var(--border2);background:var(--s3);color:var(--muted);border-radius:4px;padding:1px 7px;font-size:11px;cursor:pointer;}
.job-cancel:hover{color:var(--red);border-color:var(--red);}
.job-bar{height:6px;background:var(--border2);border-radius:3px;overflow:hidden;}
.job-fill{height:100%;width:var(--w,0%);background:var(--accent);border-radius:3px;transition:width 0.4s;}
.job-fill.done{background:var(--green);}
.job-fill.failed{background:var(--red);}
.job-sub{margin-top:6px;font-size:11px;color:var(--muted);font-family:ui-monospace,Consolas,monospace;}
#splash-inner{display:inline-flex;align-items:center;gap:12px;padding:8px 16px;background:rgba(0,0,0,0.35);border-radius:6px;}
#splash-text{color:#fff;font-size:13px;letter-spacing:1px;text-transform:uppercase;font-family:ui-monospace,Consolas,monospace;}
#splash .spinner{width:22px;height:22px;border-width:3px;border-color:rgba(255,255,255,0.25);border-top-color:#fff;}
/* ─── EIN BAUWEG, ZWEI ANORDNUNGEN ──────────────────────────────────────────
   Kachel und Zeile enthalten ab 1.6.0 dieselben Felder mit denselben Klassen.
   Was sie unterscheidet, steht ausschliesslich hier: .zeile ordnet sie in
   einer Rasterzeile an, .kachel stapelt sie und macht Platz fuer das Bild.
   Eine neue Spalte braucht deshalb nur noch einen Eintrag in DF_FELDER und
   eine Breite in .zeile — nicht mehr zwei Bauwege. */

/* Gemeinsame Felder */
/* Der Stern darf beim Anschalten nicht wachsen. Deshalb dieselbe Zeichenform
   in beiden Zustaenden — das gefuellte Emoji ⭐ wird von der Schrift groesser
   gesetzt als das leere ☆ und liess die Spalte springen. */
.df-stern{cursor:pointer;user-select:none;font-size:13px;line-height:1;width:14px;display:inline-block;text-align:center;color:var(--muted);opacity:0.45;}
.df-stern.on{color:var(--gold);opacity:1;}
.df-modell{font-family:ui-monospace,Consolas,monospace;font-size:10px;font-weight:700;padding:2px 7px;border-radius:3px;text-align:center;background:color-mix(in srgb,var(--mc) 18%,transparent);color:var(--mc);white-space:nowrap;}
.df-name{font-weight:700;font-size:13px;display:inline-flex;align-items:center;gap:5px;min-width:0;max-width:100%;}
.df-name-txt{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;min-width:0;}
.fw-pill{flex-shrink:0;display:inline-block;padding:1px 7px;border-radius:10px;font-size:9px;font-weight:800;letter-spacing:0.3px;background:var(--accent);color:#fff;white-space:nowrap;cursor:help;}
.df-online{flex-shrink:0;display:inline-block;padding:1px 7px;border-radius:3px;font-size:9px;font-weight:800;letter-spacing:0.3px;white-space:nowrap;}
.df-online.on{background:var(--gdim);color:var(--green);}
.df-online.off{background:rgba(255,80,80,0.15);color:var(--red);}
.df-repair{flex-shrink:0;display:inline-block;padding:1px 7px;border-radius:10px;font-size:9px;font-weight:800;letter-spacing:0.3px;white-space:nowrap;background:rgba(255,140,0,0.18);color:var(--orange,#ff8c00);}
.fw-card .fw-txt{font-size:13px;color:var(--muted);line-height:1.6;margin-bottom:14px;}
.fw-status{font-size:12px;color:var(--muted);}
.fw-liste{margin-top:12px;display:flex;flex-direction:column;gap:8px;}
.fw-item{display:flex;align-items:center;justify-content:space-between;gap:12px;padding:8px 12px;background:var(--s2);border:1px solid var(--border);border-radius:6px;}
.fw-item-info{display:flex;flex-direction:column;min-width:0;}
.fw-item-name{font-weight:700;font-size:13px;}
.fw-item-detail{font-size:11px;color:var(--muted);font-family:ui-monospace,Consolas,monospace;}
.fw-roh{margin-top:12px;}
.fw-roh details{font-size:11px;color:var(--muted);}
.fw-roh summary{cursor:pointer;user-select:none;}
.fw-roh pre{margin-top:8px;padding:10px;background:var(--s3);border:1px solid var(--border);border-radius:5px;overflow-x:auto;font-family:ui-monospace,Consolas,monospace;font-size:10px;line-height:1.5;white-space:pre-wrap;word-break:break-all;}
.ql-name{display:inline-flex;align-items:center;gap:6px;min-width:0;}
.df-datei{font-family:ui-monospace,Consolas,monospace;font-size:11px;color:var(--text);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.df-datei.fehler{color:var(--red);}
.df-fortschritt{display:flex;align-items:center;gap:6px;min-width:0;}
.df-balken{flex:1;height:6px;background:var(--border2);border-radius:3px;overflow:hidden;min-width:40px;display:block;}
.df-balken-fuell{display:block;height:100%;width:var(--w,0%);border-radius:3px;background:var(--accent);transition:width 0.5s;}
.df-balken-fuell.running{background:var(--green);}
.df-balken-fuell.pause{background:var(--gold);}
.df-balken-fuell.failed{background:var(--red);}
.df-balken-fuell.finish{background:var(--muted);}
.df-prozent{font-size:10px;font-family:ui-monospace,Consolas,monospace;color:var(--muted);white-space:nowrap;}
.df-leer{font-size:10px;color:var(--muted);}
.df-tag.fehler{font-size:10px;font-weight:700;color:var(--red);}
.df-rest{font-size:11px;font-family:ui-monospace,Consolas,monospace;color:var(--muted);white-space:nowrap;}
.df-ip{font-family:ui-monospace,Consolas,monospace;font-size:11px;color:var(--muted);white-space:nowrap;}
.df-status{font-size:9px;font-weight:700;letter-spacing:0.6px;padding:1px 6px;border-radius:3px;white-space:nowrap;}
.df-status:empty{display:none;}
.df-status.running{color:var(--green);background:var(--gdim);}
.df-status.pause{color:var(--gold);background:rgba(255,215,0,0.12);}
.df-status.failed{color:var(--red);background:rgba(255,80,80,0.12);}
.df-status.finish{color:var(--muted);background:var(--s3);}
.df-item.kachel .df-pillen{align-items:center;}
.df-pillen .df-online,.df-pillen .df-status,.df-pillen .df-modell,.df-pillen .cam-pill{height:18px;font-size:10px;font-weight:normal;letter-spacing:0.2px;line-height:1;padding:0 7px;border-radius:3px;display:inline-flex;align-items:center;box-sizing:border-box;}
.df-pillen .df-online-wrap{display:inline-flex;align-items:center;}

/* Anordnung 1: Zeile */
.df-item.zeile,.list-thead{display:grid;grid-template-columns:12px 20px 52px minmax(160px,1.4fr) 82px minmax(0,1fr) 150px 70px 110px;align-items:center;gap:8px;padding:6px 12px;}
.list-thead span.sortbar{cursor:pointer;user-select:none;white-space:nowrap;}
.list-thead span.sortbar:hover{color:var(--text);}
.list-thead span.aktiv{color:var(--accent);}
.sync-sortbar{display:flex;gap:16px;padding:6px 10px;border-bottom:1px solid var(--border);margin-bottom:6px;font-size:10px;text-transform:uppercase;letter-spacing:0.6px;color:var(--muted);font-family:ui-monospace,Consolas,monospace;}
.sync-sortbar span{cursor:pointer;user-select:none;}
.sync-sortbar span:hover{color:var(--text);}
.sync-sortbar span.aktiv{color:var(--accent);}
.grid-sortbar{display:flex;flex-wrap:wrap;align-items:center;gap:7px;padding:8px 10px;font-size:10px;text-transform:uppercase;letter-spacing:0.6px;color:var(--muted);font-family:ui-monospace,Consolas,monospace;background:var(--seg-bg);margin-bottom:8px;}
.grid-sortbar .gsb-label{opacity:0.7;align-self:center;}
.grid-sortbar span.sortbar{cursor:pointer;user-select:none;border:1px solid var(--border);border-radius:5px;padding:5px 11px;background:var(--s1);}
.grid-sortbar span.sortbar:hover{color:var(--text);border-color:var(--border2);}
.grid-sortbar span.aktiv{color:var(--accent);border-color:var(--accent);background:var(--adim);}
.df-item.zeile{border-bottom:1px solid var(--border);font-size:12px;background:var(--seg-bg);cursor:pointer;}
.list-thead{padding:5px 12px;font-size:10px;font-family:ui-monospace,Consolas,monospace;color:var(--muted);text-transform:uppercase;letter-spacing:0.8px;border-bottom:1px solid var(--border);background:var(--s2);}
.df-item.zeile:hover{background:var(--s2);}
.df-item.zeile.fav{box-shadow:inset 3px 0 0 var(--gold);}
.df-item.zeile.hat-fehler{box-shadow:inset 3px 0 0 var(--red);}
.df-item.zeile.auf{background:var(--s2);}

/* Anordnung 2: Kachel */
.df-item.kachel{display:flex;flex-direction:column;}
.df-item.kachel .df-kopf{display:flex;align-items:center;gap:7px;padding:7px 9px 0;}
.df-item.kachel .df-pillen{display:flex;align-items:center;gap:5px;padding:5px 9px;flex-wrap:wrap;}
.df-item.kachel .df-balken.breit{display:none;}
.df-item.kachel .df-fuss{display:flex;flex-direction:column;gap:2px;padding:5px 9px 7px;min-height:40px;}
.df-item.kachel .df-fuss .df-fuss-r1{min-height:15px;display:flex;align-items:center;}
.df-item.kachel .df-fuss .df-fuss-r2{display:flex;align-items:center;gap:8px;min-height:16px;}
.df-item.kachel .df-fuss .df-fuss-r2 .df-rest{flex:none;}
.df-item.kachel .df-fuss .df-fuss-r2 .df-fortschritt{flex:1;min-width:0;justify-content:flex-end;}
/* Im Fuss zaehlt nur noch die Zahl — der Balken steht jetzt eine Zeile hoeher,
   direkt unter dem Bild. */
.df-item.kachel .df-fuss .df-balken{display:block;}
.df-item.kachel .df-name{flex:1;}
.df-item.kachel .df-fortschritt{min-width:90px;}
/* Erreichbarkeit */
.ds-box{margin-top:6px;padding:8px;background:var(--s3);border:1px solid var(--border);border-radius:5px;}
.ds-box.hidden{display:none;}
.ds-zeile{display:flex;align-items:center;gap:8px;margin-bottom:6px;flex-wrap:wrap;}
.ds-lbl{font-size:11px;color:var(--muted);width:60px;flex:none;}
.ds-feld{flex:1;min-width:140px;background:var(--s2);border:1px solid var(--border2);border-radius:4px;color:var(--text);padding:4px 6px;font-size:11px;font-family:ui-monospace,Consolas,monospace;}
.ds-check{font-size:11px;color:var(--muted);display:flex;align-items:center;gap:4px;cursor:pointer;}
.ds-hinweis{font-size:10px;color:var(--muted);line-height:1.45;margin-bottom:6px;}
.ds-map-titel{font-size:10px;text-transform:uppercase;letter-spacing:0.6px;color:var(--muted);margin:4px 0 3px;}
.ul-liste{max-height:260px;overflow-y:auto;margin-top:8px;}
.ul-zeile{display:grid;grid-template-columns:110px 130px minmax(0,1fr) 70px 150px;gap:8px;align-items:center;padding:3px 6px;border-bottom:1px solid var(--border);font-size:11px;}
.ul-zeile:last-child{border-bottom:none;}
.ul-zeile>span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.ul-zeit,.ul-groesse{font-family:ui-monospace,Consolas,monospace;color:var(--muted);}
.ul-datei{font-family:ui-monospace,Consolas,monospace;}
.ul-ergebnis{color:var(--green);}
.ul-zeile.fehlgeschlagen .ul-ergebnis{color:var(--red);}
.rc-row{display:grid;grid-template-columns:minmax(120px,170px) max-content minmax(0,1fr);gap:12px;align-items:center;padding:5px 8px;border-bottom:1px solid var(--border);font-size:12px;}
.rc-row>span{min-width:0;overflow:hidden;text-overflow:ellipsis;}
.rc-row:last-child{border-bottom:none;}
.rc-ports{display:flex;gap:5px;flex:none;white-space:nowrap;overflow:visible;}
.rc-p{font-family:ui-monospace,Consolas,monospace;font-size:10px;padding:1px 5px;border-radius:3px;border:1px solid var(--border);}
.rc-p.auf{border-color:var(--green);color:var(--green);}
.rc-p.zu{border-color:var(--red);color:var(--red);}
.rc-urteil{color:var(--muted);white-space:normal;overflow-wrap:anywhere;}
.rc-urteil.gut{color:var(--green);}
.rc-urteil.warn{color:var(--gold);}
.rc-urteil.schlecht{color:var(--red);}
.busy-badge{display:none;align-items:center;gap:6px;font-size:11px;color:var(--muted);}
@keyframes df-run{0%,100%{outline-color:transparent;}50%{outline-color:var(--accent);}}
.sbtn.res-btn.working{outline:2px solid transparent;outline-offset:2px;animation:df-run 1.6s ease-in-out infinite;}
.busy-badge.shown{display:inline-flex;}
.tile-loading{position:absolute;inset:0;display:flex;align-items:center;justify-content:center;background:rgba(0,0,0,0.25);pointer-events:none;}
html[data-theme="light"] .tile-loading{background:rgba(255,255,255,0.35);}
.skeleton{border:1px dashed var(--border2);border-radius:6px;padding:18px;text-align:center;color:var(--muted);font-size:12px;line-height:1.7;}
.skeleton .spinner{margin-bottom:8px;}
/* Gegenstuecke zu den Regeln, die per Stylesheet ausblenden. Ohne diese bleibt
   eine Ansicht auf display:none, auch wenn "hidden" entfernt wurde. */
.view.shown{display:block;}
#view-uebersicht.shown{display:flex;flex-direction:column;flex:1;}
.sd-files.shown{display:block;}
.ir.shown{display:flex;}
.clickable{cursor:pointer;}
.grow{flex:1;}
.right{text-align:right;}
.push-right{margin-left:auto;}
.mt5{margin-top:5px;}.mt8{margin-top:8px;}.mt10{margin-top:10px;}.mt12{margin-top:12px;}
.ml2{margin-left:2px;}.ml4{margin-left:4px;}.ml6{margin-left:6px;}.mr6{margin-right:6px;}.mx8{margin:0 8px;}
.w70{width:70px;}.minw70{min-width:70px;}
.row-8{display:flex;gap:8px;align-items:center;flex-wrap:wrap;}
.row-8-tight{display:flex;gap:8px;}
.row-8-mt{display:flex;gap:8px;margin-top:10px;align-items:center;}
.row-10{display:flex;gap:10px;align-items:center;flex-wrap:wrap;}
.row-10-mb{display:flex;align-items:center;gap:10px;margin-bottom:6px;}
.col-grow{display:flex;flex-direction:column;flex:1;}
.btn-row{display:flex;gap:6px;margin-top:12px;flex-wrap:wrap;align-items:center;}
.btn-inline{display:flex;gap:6px;align-items:center;}
.btn-sm{height:26px;padding:0 10px;font-size:11px;}
.btn-xs{font-size:10px;height:24px;}
.btn-add{height:36px;align-self:end;}
.btn-danger{color:var(--red);border-color:rgba(255,84,112,0.5);}
.sbtn-xs{font-size:10px;padding:0 8px;}
.brand{padding:0 16px;border-right:1px solid var(--border-menu);display:flex;align-items:center;gap:10px;height:100%;}
.header-right-inner{padding:0;gap:0;display:flex;align-items:center;height:100%;}
.tab-header{border-bottom:none;height:100%;padding:0 15px;}
.tab-sub{height:30px;padding:0 16px;font-size:12px;border-bottom:none;}
.tab-sub-sm{height:30px;padding:0 14px;font-size:11px;border-bottom:none;}
.tab-sub-sep{height:30px;padding:0 14px;font-size:11px;border-bottom:none;border-left:1px solid var(--border);}
.setup-icon{color:inherit;text-shadow:none;margin-right:4px;}
.mode-group{margin-left:auto;display:flex;align-items:center;gap:4px;}
.mode-label{font-size:10px;color:var(--muted);font-family:ui-monospace,Consolas,monospace;letter-spacing:0.5px;}
.fill-media{width:100%;height:100%;object-fit:cover;display:block;}
/* Winziger Ladekreis oben links, laeuft genau so lange wie das Snapshot-
   Intervall (--takt) — bei 2, 5 oder 10 Sekunden entsprechend. */
.snap-ring{position:absolute;top:3px;left:3px;width:10px;height:10px;border-radius:50%;border:1.5px solid rgba(255,255,255,0.25);border-top-color:rgba(255,255,255,0.9);animation:snap-dreh var(--takt,2s) linear infinite;pointer-events:none;}
@keyframes snap-dreh{to{transform:rotate(360deg);}}
.snap-stamp{position:absolute;bottom:2px;right:4px;font-size:9px;color:var(--muted);font-family:ui-monospace,Consolas,monospace;}
.lead{font-size:12px;color:var(--muted);line-height:1.6;margin:4px 0 10px;}
.setup-info{font-size:12px;line-height:1.6;color:var(--muted);margin:4px 0 0;}
.muted-note{font-size:12px;color:var(--muted);}
.count-note{color:var(--muted);font-size:13px;font-weight:400;}
.empty-note{text-align:center;padding:40px;color:var(--muted);font-size:13px;}
.hint-xs{font-size:10px;opacity:0.6;}
.sep-note{font-size:9px;color:var(--muted);white-space:nowrap;margin:0 6px;}
.ok-note{font-size:11px;color:var(--green);margin-bottom:4px;}
.tag-green{color:var(--green);font-size:11px;font-weight:700;}
.tag-orange{color:var(--orange);}
.frow-add{grid-template-columns:110px 1fr 1fr 1fr 1fr auto;}
.search-field{flex:1;min-width:220px;height:32px;}
.sel-bar{margin-top:10px;gap:10px;align-items:center;flex-wrap:wrap;display:flex;}
.card-inset{margin-top:10px;padding:12px;}
.bar-track{height:8px;background:var(--s3);border-radius:4px;overflow:hidden;}
.bar-track-slim{height:5px;background:var(--s3);border-radius:3px;overflow:hidden;margin-top:9px;}
.bar-fill{height:100%;width:var(--w,0%);background:var(--accent);transition:width 0.3s;}
.progress-wrap{margin-top:10px;}
.progress-track{height:8px;background:var(--s3);border-radius:4px;overflow:hidden;}
.progress-fill{height:100%;width:var(--w,0%);background:var(--accent);transition:width 0.3s;}
.progress-text{margin:6px 0 0;font-size:12px;color:var(--muted);font-family:ui-monospace,Consolas,monospace;}
.release-notes{margin:10px 0 0;font-size:12px;color:var(--muted);line-height:1.6;white-space:pre-wrap;}
.modal-backdrop{position:fixed;inset:0;z-index:9500;background:rgba(0,0,0,0.72);align-items:center;justify-content:center;display:flex;}
.modal-box{background:var(--s2);border:1px solid var(--border2);border-radius:10px;max-width:520px;width:calc(100% - 40px);padding:22px 24px;box-shadow:0 20px 60px rgba(0,0,0,0.6);}
.ping-overlay{position:fixed;inset:0;z-index:9600;background:rgba(0,0,0,0.55);display:flex;align-items:center;justify-content:center;}
.ping-box{max-width:440px;}
.ping-title{font-weight:800;font-size:15px;margin-bottom:10px;display:flex;align-items:flex-start;justify-content:space-between;gap:10px;}
.ping-msg{font-size:12px;margin-bottom:12px;}
.ping-msg.rot{color:var(--red);} .ping-msg.gruen{color:var(--green);}
.ping-row{display:flex;justify-content:space-between;gap:14px;padding:6px 0;border-bottom:1px solid var(--border);font-size:13px;}
.ping-was{color:var(--muted);}
.ping-urteil{margin-top:12px;font-size:12px;color:var(--text);}
.dd-close{cursor:pointer;color:var(--muted);font-size:15px;line-height:1;padding:2px 6px;border-radius:5px;flex-shrink:0;}
.dd-close:hover{background:var(--s2);color:var(--text);}
.ping-hint{margin-top:8px;font-size:12px;color:var(--orange,#ff8c00);background:rgba(255,140,0,0.1);border-radius:5px;padding:7px 9px;}
.modal-title{margin:0 0 6px;font-size:17px;color:var(--text);}
.modal-lead{margin:0 0 16px;font-size:13px;line-height:1.6;color:var(--muted);}
.modal-list{display:flex;flex-direction:column;gap:8px;margin-bottom:16px;}
.modal-progress{margin-bottom:16px;}
.modal-error{margin:0 0 14px;font-size:12px;line-height:1.6;color:var(--red);word-break:break-word;}
.modal-actions{display:flex;gap:8px;justify-content:flex-end;flex-wrap:wrap;}
.modal-footnote{margin:14px 0 0;font-size:11px;line-height:1.6;color:var(--muted);}
.ql-name{font-size:13px;font-weight:600;flex:1;}
.ql-ip{font-size:11px;color:var(--muted);}
.ql-row input{width:100%;}
.set-inline{display:flex;align-items:center;gap:10px;flex-wrap:wrap;}
.col-range{width:200px;}
.col-label{font-family:ui-monospace,Consolas,monospace;font-size:13px;min-width:34px;}
.repo-input{width:320px;max-width:100%;}
.req{color:var(--red);margin-left:2px;font-weight:700;}
.comp-list{display:flex;flex-direction:column;gap:10px;margin-top:12px;}
.comp-updates{margin-top:10px;}
.comp-item{display:flex;gap:10px;align-items:flex-start;font-size:12px;}
.comp-dot{width:8px;height:8px;border-radius:50%;margin-top:5px;flex:none;background:var(--muted);}
.comp-dot.ok{background:var(--green);}
.comp-dot.missing{background:var(--red);}
.comp-body{min-width:0;flex:1;}
.comp-title{color:var(--text);}
.comp-purpose{color:var(--muted);}
.comp-detail{color:var(--muted);font-family:ui-monospace,Consolas,monospace;font-size:11px;word-break:break-all;}
.comp-sub-row{display:flex;gap:8px;align-items:center;margin-top:6px;font-size:11px;}
.comp-sub-label{color:var(--muted);}
.yaml-link{color:var(--accent);font-family:ui-monospace,Consolas,monospace;word-break:break-all;text-decoration:none;}
.yaml-link:hover{text-decoration:underline;}
.ams-dot.multi{background:var(--grad);}
.ams-dot[style*="--tray"]{background:var(--tray);}
.sync-dot{background:var(--mc);width:8px;height:8px;border-radius:50%;flex-shrink:0;}
.mdot-color{background:var(--mc);}
.mbadge-color{background:color-mix(in srgb,var(--mc) 15%,transparent);color:var(--mc);}
.fav-toggle{font-size:12px;opacity:var(--o,1);cursor:pointer;}
.bar-fill.warn{background:var(--orange);}
.fs-grid{display:grid;grid-template-columns:repeat(var(--cols),1fr);grid-auto-rows:var(--cellh);gap:0;}
.fs-grid-tiles{display:grid;grid-template-columns:repeat(var(--cols),1fr);grid-auto-rows:var(--cellh);gap:2px;height:100vh;box-sizing:border-box;}
.tile-live-badge.dimmed{background:var(--muted2)!important;animation:none!important;color:#aaa!important;}
.list-row.row-error{background:rgba(255,51,85,0.08);border-left:3px solid var(--red);}
.list-row.row-pause{background:rgba(255,165,0,0.06);border-left:3px solid var(--orange);}
.list-row.row-done{background:rgba(0,199,80,0.06);}
/* Solange alles laeuft, ist das Zahnrad unauffaellig — es meldet sich erst,
   wenn etwas fehlt. Ein dauerhaft leuchtendes Symbol traegt keine Information. */
/* Solange nichts im Netz geschieht, muss das unuebersehbar sein — sonst sucht
   man den Fehler bei den Druckern. */
.netz-aus{color:var(--red);margin-right:5px;font-size:13px;animation:netz-puls 1.8s ease-in-out infinite;}
.netz-aus.hidden{display:none;}
@keyframes netz-puls{0%,100%{opacity:1;}50%{opacity:0.35;}}
body.netz-pause .df-item.kachel{opacity:0.55;}
#setup-icon{color:inherit;text-shadow:none;}
#setup-icon.fehlt{color:var(--red);text-shadow:0 0 8px var(--red);}
.sd-space-fill.failed{background:var(--red);}
.fsv-overlay{position:fixed;inset:0;z-index:9000;background:#000;cursor:pointer;display:grid;grid-template-columns:repeat(var(--cols,9),1fr);gap:0;overflow-y:auto;}
body.broadcast-mode>*:not(#bc-root){display:none !important;}
#bc-root{display:none;}
body.broadcast-mode{overflow:hidden;background:#000;}
body.broadcast-mode #bc-root{display:block;position:fixed;inset:0;background:#000;overflow:auto;z-index:10000;}
body.broadcast-mode.bc-noclick #bc-root{pointer-events:none;}
#bc-grid{display:grid;grid-template-columns:repeat(var(--cols,9),1fr);gap:2px;background:#0a0a0a;padding:2px;min-height:100vh;box-sizing:border-box;}
#bc-grid .fsv-cell{aspect-ratio:16/9;}
.fs-view{display:block;background:#000;position:fixed;inset:0;z-index:500;overflow:hidden;}
#printerGrid{grid-template-columns:repeat(var(--cols,6),1fr);}
.tile-video.fading{opacity:0;}
.tile-video{transition:opacity 0.2s;}
.sd-arrow{transition:transform 0.15s;}
.sd-arrow.open{transform:rotate(180deg);}
.diag-box{margin-top:12px;border:1px solid var(--border2);border-radius:6px;padding:10px;background:var(--s2);}
.diag-text{margin:8px 0 0;font-family:ui-monospace,Consolas,monospace;font-size:11px;line-height:1.5;color:var(--text);white-space:pre-wrap;word-break:break-word;max-height:420px;overflow-y:auto;}
.diag-warn{color:var(--red);font-weight:700;}
.copyright-card{text-align:center;}
.copyright-card p{margin:0;font-size:11px;color:var(--muted);}
.af-unit{margin-bottom:8px;}
.af-kopf{font-size:10px;font-weight:700;color:var(--accent);margin-bottom:3px;}
.af-fach{display:flex;align-items:center;gap:5px;margin-bottom:3px;font-size:10px;}
.af-nr{width:12px;flex:none;text-align:center;color:var(--muted);}
.af-typ{width:38px;flex:none;font-family:ui-monospace,Consolas,monospace;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.af-editor{display:flex;align-items:center;gap:3px;flex:1;min-width:0;}
.af-sel{flex:1;min-width:0;background:var(--s3);border:1px solid var(--border2);border-radius:3px;color:var(--text);font-size:10px;padding:1px 2px;}
.af-farbe{width:22px;height:18px;flex:none;border:1px solid var(--border2);border-radius:3px;background:none;padding:0;cursor:pointer;}
.af-set{flex:none;background:var(--s2);border:1px solid var(--border2);border-radius:3px;color:var(--green);cursor:pointer;font-size:11px;padding:0 5px;}
.af-set:hover{border-color:var(--green);}
.af-fach .ams-dot{flex:0 0 11px;}
.set-filter{display:flex;gap:12px;flex-wrap:wrap;}
.set-check{display:flex;align-items:center;gap:5px;font-size:12px;color:var(--muted);cursor:pointer;}
.dd-cols{display:grid;grid-template-columns:1fr 1fr;gap:12px;align-items:stretch;}
.dd-cols .dd-c{min-width:0;}
@media (max-width:520px){.dd-cols{grid-template-columns:1fr;}}
.dd-blocks{display:flex;flex-wrap:wrap;gap:12px;align-items:stretch;}
.dd-blocks .dd-block{flex:1 1 240px;min-width:0;background:var(--s2);border:1px solid var(--border);border-radius:6px;padding:8px 10px;}
.dd-blocks .dd-block-title{font-size:10px;text-transform:uppercase;letter-spacing:0.6px;color:var(--muted);margin-bottom:6px;font-weight:700;}
.scan-cell{position:relative;overflow:hidden;background:#000;}
.fsv-cell{box-sizing:border-box;border:1px solid var(--border);aspect-ratio:auto;height:stretch;}
.fsv-cell.running{border-color:var(--green);}
.fsv-cell.pause{border-color:var(--gold);}
.fsv-cell.failed{border-color:var(--red);}
.fsv-cell.finish{border-color:var(--muted);}
.fsv-cell.idle{border-color:#1e2730;}
.fsv-cell .fill-media{object-fit:unset;height:unset;}
.fsv-head{position:absolute;top:0;left:0;z-index:4;max-width:88%;display:flex;flex-direction:column;gap:1px;align-items:flex-start;}
.fsv-name{max-width:100%;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;background:rgba(0,0,0,0.55);color:#fff;font-size:11px;font-weight:700;padding:2px 7px;border-bottom-right-radius:5px;font-family:ui-monospace,Consolas,monospace;}
.fsv-sub{background:rgba(0,0,0,0.45);color:#fff;font-size:10px;padding:1px 7px;border-bottom-right-radius:5px;font-family:ui-monospace,Consolas,monospace;}
.fs-spin{position:absolute;inset:0;display:flex;align-items:center;justify-content:center;background:rgba(2,4,6,0.55);z-index:2;pointer-events:none;}
.fsv-bar{position:absolute;left:0;right:0;bottom:0;height:6px;background:rgba(0,0,0,0.45);z-index:3;}
.fsv-bar-fill{display:block;height:100%;width:var(--w,0%);background:var(--accent);transition:width 0.5s;}
.fsv-bar-fill.running{background:var(--green);}
.fsv-bar-fill.pause{background:var(--gold);}
.fsv-bar-fill.failed{background:var(--red);}
.fsv-bar-fill.finish{background:var(--muted);}
.scan-cell-fixed{position:relative;aspect-ratio:16/9;background:#000;overflow:hidden;}

</style>
</head>
<body>
<div id="bc-root"><div id="bc-grid" class="fsv-grid"></div></div>
<div id="ueber" class="hidden" onclick="closeSplash()">
  <div id="ueber-inhalt">
    <div id="ueber-text">
      <img src="/logo.svg?v=__BUILD_ID__" alt="Druckerfarm" class="ueber-logo">
      <p class="ueber-suffix">Switzerland</p>
      <p class="ueber-copy" id="ueber-copy"></p>
      <p class="ueber-version" id="ueber-version"></p>
      <p class="ueber-version" id="ueber-build">Build __BUILD_ID__</p>
      <p class="ueber-hinweis">Klicken zum Schließen</p>
    </div>
  </div>
</div>
<div id="splash">
  <div id="splash-scharf"></div>
  <div id="splash-mitte">
    <img src="/logo.svg?v=__BUILD_ID__" alt="Druckerfarm" class="splash-logo">
    <p class="splash-suffix">Switzerland</p>
    <div id="splash-inner">
      <span class="spinner"></span>
      <span id="splash-text">Druckerfarm startet …</span>
    </div>
    <p class="splash-copy">© Brickshouse GmbH · Druckerfarm.ch</p>
    <p class="splash-ver">Version __APP_VERSION__</p>
    <p class="splash-ver" id="splash-build">Build __BUILD_ID__</p>
  </div>
</div>
<div id="lang-picker" class="lang-picker" style="display:none;">
  <div class="lp-box">
    <img src="/logo.svg?v=__BUILD_ID__" alt="Druckerfarm" class="lp-logo">
    <h2 class="lp-title">Sprache wählen<br><span class="lp-sub">Choose language</span></h2>
    <div class="lp-btns">
      <button class="lp-btn" onclick="pickLanguage('de')">Deutsch</button>
      <button class="lp-btn" onclick="pickLanguage('en')">English</button>
      <button class="lp-btn" onclick="pickLanguage('zh')">中文</button>
      <button class="lp-btn" onclick="pickLanguage('es')">Español</button>
    </div>
  </div>
</div>

<header>
  <!-- Logo + App name -->
  <div class="logo brand clickable" onclick="showSplash()" title="Über Druckerfarm">
    <img src="/logo.svg?v=__BUILD_ID__" alt="Druckerfarm" class="app-logo">
    <span class="app-suffix">Switzerland</span>
  </div>
  <!-- Main nav left -->
  <div class="header-tabs">
    <div class="tab active" id="tab-uebersicht" onclick="sv('uebersicht')">Übersicht</div>
    <div class="tab tab-header" id="tab-sync" onclick="sv('sync')"><span id="tab-sync-lbl2">⟳ File Sync</span></div>
  </div>
  <!-- Spacer -->
  <div class="grow"></div>
  <!-- Right side -->
  <div class="header-right header-right-inner">
    <div class="tab tab-header" id="tab-jobs" onclick="sv('jobs')" title="Aufträge">⚙ Aufträge<span id="jobs-badge" class="jobs-badge" style="display:none;">0</span></div>
    <div class="tab tab-header" onclick="sv('import')" id="tab-import">
      <span id="tab-import-lbl">+ Drucker</span>
    </div>
    <div class="tab tab-header" onclick="sv('setup')" id="tab-setup">
      <span id="netz-aus" class="netz-aus hidden" title="Netzwerkverkehr ist angehalten">🚫</span>
      <span id="setup-icon" class="setup-icon">⚙</span><span id="tab-setup-lbl">Einstellungen</span>
    </div>
    <div class="user-chip" id="user-chip" title="Angemeldeter Benutzer" style="display:none;">
      <span class="user-name" id="user-name">–</span>
      <span class="user-role" id="user-role"></span>
      <button class="hbtn btn-sm" onclick="doLogout()" data-i18n="logout" title="Abmelden">Abmelden</button>
    </div>
  </div>
</header>
<div id="praesenz-banner" class="praesenz-banner" style="display:none;"></div>

<main>
<div class="view" id="view-uebersicht">
<!-- Sub nav: Grid / Liste + Slider -->
<div class="sub-nav" id="sub-nav">
  <div class="tab active tab-sub" id="tab-grid" onclick="setView('grid')"><span id="tab-grid-lbl">⊞ Grid</span></div>
  <div class="tab tab-sub" id="tab-list" onclick="setView('list')"><span id="tab-list-lbl">≡ Liste</span></div>
  <div class="push-right"></div>
  <div class="tab tab-sub-sep" id="tab-fsc" onclick="setView('fsc')" data-i18n="fullscreen">⊡ Fullscreen</div>
  <div class="tab tab-sub-sm" id="tab-fsv" onclick="setView('fsv')" data-i18n="fullscreenVideo">⛶ Fullscreen Video Only</div>
</div>
<div class="stream-bar" id="stream-bar">
    <span class="stream-bar-label" data-i18n="streams">Streams</span>
    <div class="stream-divider"></div>
    <button class="sbtn s-play" onclick="startAll()" id="btn-startall"></button>
    <button class="sbtn s-stop" onclick="stopAll()" id="btn-stopall"></button>
    <div class="stream-divider"></div>
    <div class="mode-group">
      <span class="mode-label" data-i18n="profile">Profile</span>
      <button class="sbtn res-btn res-active sbtn-xs" id="btn-snap2"  onclick="setSnapMode(2)">📷 2s</button>
      <button class="sbtn res-btn sbtn-xs" id="btn-snap5"  onclick="setSnapMode(5)">📷 5s</button>
      <button class="sbtn res-btn sbtn-xs" id="btn-snap10" onclick="setSnapMode(10)">📷 10s</button>
      <div class="stream-divider"></div>
      <button class="sbtn res-btn sbtn-xs" id="btn-720p"  onclick="setVideoMode('720p')">720p</button>
      <button class="sbtn res-btn sbtn-xs" id="btn-1080p" onclick="setVideoMode('1080p')">1080p</button>
    </div>
  </div>
<div class="top-bar" id="top-bar">
    <span class="filter-label" id="lbl-filter"></span>
    <div class="stream-divider"></div>
    <div class="stat-pill active-all" id="pill-all" onclick="setPillFilter('all')">
      <div class="sl" id="lbl-total"></div><div class="sv ct" id="st0">0</div>
    </div>
    <div class="stat-pill" id="pill-online" onclick="setPillFilter('online')"><div class="sl" id="lbl-online-pill"></div><div class="sv sv-green" id="st-online">0</div></div>
    <div class="stat-pill" id="pill-offline" onclick="setPillFilter('offline')"><div class="sl" data-i18n="pillOffline">Offline</div><div class="sv sv-red" id="st-offline">0</div></div>
    <div class="stat-pill" id="pill-paused" onclick="setPillFilter('paused')"><div class="sl" data-i18n="pillPaused">Pausiert</div><div class="sv sv-amber" id="st-paused">0</div></div>
    <div class="stat-pill" id="pill-favs" onclick="setPillFilter('favs')"><div class="sl" data-i18n="favs">Favoriten</div><div class="sv sv-gold" id="st-favs">0</div></div>
    <div class="stat-pill" id="pill-fertig" onclick="setPillFilter('fertig')"><div class="sl" data-i18n="pillFertig">Fertig</div><div class="sv sv-green" id="st-fertig">0</div></div>
    <span id="model-pills" class="pill-host"></span>
    <input class="search-input" id="searchInput" oninput="flt(this.value)">
</div>
<div class="view" id="view-grid">
  <div id="grid-sortbar" class="grid-sortbar"></div>
  <div id="printerGrid"></div>
</div>

<div class="view" id="view-list">
  <div id="printerList"></div>
</div>
</div>

<div class="view" id="view-import">
  <div class="ip">
    <div class="pc">
      <h2>📡 Drucker im Netzwerk suchen</h2>
      <p class="lead">
        Sucht per SSDP im lokalen Netz. Gefundene Geräte liefern Modell, Name, IP und Seriennummer —
        den Zugangscode musst du selbst eintragen, den gibt der Drucker nicht heraus.
      </p>
      <div class="row-8">
        <button class="hbtn primary" id="btn-scan" onclick="scanNetwork()">📡 Netzwerk nach Druckern durchsuchen</button>
        <span id="scan-status" class="note"></span>
      </div>
      <div id="scan-results" class="mt12"></div>
    </div>
    <div class="pc">
      <h2 id="imp-csv-title"></h2>
      <p id="imp-csv-desc"></p>
      <div class="dz" id="dz" onclick="document.getElementById('fi').click()">
        <p id="dz-text"></p>
        <input type="file" id="fi" accept=".csv,.txt" onchange="rdF(this)">
      </div>
      <div class="ia">
        <button class="hbtn" onclick="setShown(document.getElementById('ir'), false)" id="btn-clear"></button>
        <div class="ir" id="ir"></div>
      </div>
    </div>
    <div class="pc">
      <h2 id="imp-manual-title"></h2>
      <div class="frow frow-add">
        <div class="fg"><label id="lbl-model"></label><input id="mm" type="text" oninput="this.value=this.value.toUpperCase()"></div>
        <div class="fg"><label id="lbl-name"></label><input id="mn" type="text"></div>
        <div class="fg"><label><span id="lbl-ip"></span><span class="req">*</span></label><input id="mi" type="text" placeholder="192.168.1.100"></div>
        <div class="fg"><label><span id="lbl-code"></span><span class="req">*</span></label><input id="mc" type="text" placeholder="abcd1234"></div>
        <div class="fg"><label><span id="lbl-serial"></span><span class="req">*</span></label><input id="ms" type="text" placeholder="01P00C123456789"></div>
        <button class="hbtn primary btn-add" onclick="addM()" id="btn-add"></button>
      </div>
      <div class="frow frow-add-opts">
        <label class="cbx"><input type="checkbox" id="add-kamera-aus"><span data-i18n="addOptKamera">Kamera aus</span></label>
        <label class="cbx"><input type="checkbox" id="add-blink-aus"><span data-i18n="addOptBlink">Blinken aus</span></label>
        <label class="cbx"><input type="checkbox" id="add-repair"><span data-i18n="addOptRepair">In Reparatur</span></label>
      </div>
    </div>
    <div class="pc fw-card" id="fw-card">
      <h2 data-i18n="fwTitle">🔄 Drucker-Software-Updates</h2>
      <p class="fw-txt" data-i18n="fwDesc">Fragt jeden verbundenen Drucker, ob neue Firmware bereitliegt — löst nichts aus, sucht nur. Das Ergebnis erscheint als Marke „⬆ Update" hinter dem Namen in der Liste unten und bleibt, bis der Drucker die Fassung installiert hat.</p>
      <div class="row-8">
        <button class="hbtn primary fw-btn" id="fw-btn" onclick="searchPrinterUpdates()" data-i18n="fwBtn">🔄 Drucker Software Update suchen</button>
        <span class="fw-status note" id="fw-status"></span>
      </div>
      <div id="fw-liste" class="fw-liste"></div>
      <div id="fw-roh" class="fw-roh"></div>
    </div>
    <div class="pc">
      <h2><span id="imp-saved-title"></span> <span id="sc" class="count-note"></span></h2>
      <div id="ql" class="mt8 ql-wrap"></div>
      <button class="hbtn mt12" onclick="confirmDelAll()" id="btn-delall"></button>
    </div>
    <div class="pc">
      <h2 id="cam-admin-title">📷 Kameras</h2>
      <p class="fw-txt" id="cam-admin-desc"></p>
      <div id="cam-list" class="mt8"></div>
      <details class="xiaomi-box mt8" id="xiaomi-box">
        <summary id="xiaomi-sum">Xiaomi Mi-Home-Login (für Xiaomi-Kameras)</summary>
        <p class="note-sm" id="xiaomi-desc"></p>
        <div class="cam-form" id="xm-login-form">
          <input id="xm-user" placeholder="Mi-Konto (E-Mail/Telefon)" class="cam-in" autocomplete="off">
          <input id="xm-pass" type="password" placeholder="Passwort" class="cam-in" autocomplete="off">
          <button class="hbtn btn-sm" onclick="xiaomiLogin()" id="xm-btn">Anmelden</button>
        </div>
        <div id="xm-challenge" class="hidden">
          <div id="xm-captcha-wrap" class="hidden"><img id="xm-captcha-img" class="xm-captcha" alt="Captcha"></div>
          <p class="note-sm" id="xm-challenge-hint"></p>
          <div class="cam-form">
            <input id="xm-code" placeholder="Code / Captcha" class="cam-in" autocomplete="off">
            <button class="hbtn btn-sm green" onclick="xiaomiChallenge()" id="xm-code-btn">Bestätigen</button>
          </div>
        </div>
        <div class="cam-form mt8">
          <input id="xm-server" placeholder="Region (z. B. de)" class="cam-in" value="de" style="max-width:130px">
          <button class="hbtn btn-sm" onclick="xiaomiListDevices()" id="xm-list-btn">Kameras des Kontos laden</button>
        </div>
        <div id="xm-devices" class="mt8"></div>
        <pre id="xm-result" class="xm-result note-sm"></pre>
      </details>
      <div class="cam-form mt8">
        <input id="cam-name" placeholder="Name" class="cam-in">
        <input id="cam-src" placeholder="Quelle (z. B. xiaomi://…)" class="cam-in mono">
        <button class="hbtn green btn-sm" onclick="addCamera()" id="cam-add-btn">Hinzufügen</button>
      </div>
      <p class="note-sm" id="cam-hint"></p>
    </div>
    <div class="pc">
      <h2 id="imp-export-title"></h2>
      <div class="row-8-tight">
        <button class="hbtn" onclick="expCSV()" id="btn-expcsv"></button>
        <button class="hbtn green" onclick="dlYaml()" id="btn-expyaml"></button>
      </div>
    </div>
  </div>
</div>

<div class="view" id="view-setup">
  <div class="sp">
    <div class="stcard">
      <h2>🧩 Komponenten</h2>
      <p id="setup-info" class="setup-info"></p>
      <!-- comp-list wird gefuellt; der go2rtc-Unterpunkt haengt direkt an seinem Eintrag -->
      <div id="comp-list" class="comp-list"></div>
      <div id="comp-updates" class="comp-updates"></div>
      <div id="reach-box" class="diag-box hidden">
        <div class="btn-row">
          <span id="reach-sum" class="muted-note"></span>
          <button class="hbtn push-right" onclick="setShown(byId('reach-box'), false)">Schließen</button>
        </div>
        <div id="reach-list"></div>
      </div>
      <div id="diag-box" class="diag-box hidden">
        <div class="btn-row">
          <button class="hbtn" onclick="copyDiagnostics()">In die Zwischenablage</button>
          <button class="hbtn" onclick="setShown(byId('diag-box'), false)">Schließen</button>
          <span id="diag-hint" class="muted-note"></span>
        </div>
        <pre id="diag-text" class="diag-text"></pre>
      </div>
      <div class="btn-row">
        <button class="hbtn" onclick="loadComponents(true)">Status prüfen</button>
        <button class="hbtn" onclick="checkComponentUpdates(true)" id="btn-compupd">Auf Updates prüfen</button>
        <button class="hbtn green" onclick="openSetupDialog()">Fehlende nachladen</button>
        <button class="hbtn green hidden" id="btn-compupd-run" onclick="runComponentUpdate()">Komponenten aktualisieren</button>
        <button class="hbtn" onclick="showDiagnostics()">🩺 Diagnose</button>
        <button class="hbtn" onclick="checkReach()" id="btn-reach">📡 Drucker erreichbar?</button>
        <span id="comp-upd-status" class="muted-note"></span>
      </div>
    </div>

    <div class="stcard">
      <h2>🌐 Netzwerkverkehr</h2>
      <p class="setup-info">Hält alles an, was ins Netz geht — Status, Bilder, MQTT und go2rtc. Zum Beispiel während einer Wartung. Nach einem Neustart läuft wieder alles.</p>
      <div class="btn-row">
        <button class="hbtn" onclick="netzPauseUmschalten()" id="btn-netzpause">⏸ Kompletten Netzwerkverkehr pausieren</button>
      </div>
    </div>

    <div class="stcard">
      <h2>🖥 Darstellung</h2>
      <div class="set-grid">
        <div class="set-row">
          <label>Farbschema</label>
          <div class="seg">
            <button class="seg-btn" id="theme-dark"  onclick="setTheme('dark')">Dunkel</button>
            <button class="seg-btn" id="theme-light" onclick="setTheme('light')">Hell</button>
          </div>
        </div>
        <div class="set-row">
          <label>Sprache</label>
          <div class="seg">
            <button class="seg-btn" id="lang-de" onclick="setLang('de')">Deutsch</button>
            <button class="seg-btn" id="lang-en" onclick="setLang('en')">English</button>
            <button class="seg-btn" id="lang-zh" onclick="setLang('zh')">中文</button>
            <button class="seg-btn" id="lang-es" onclick="setLang('es')">Español</button>
          </div>
        </div>
        <div class="set-row">
          <label>Videos nebeneinander</label>
          <div class="set-inline">
            <input type="range" min="1" max="12" value="6" id="colR" class="col-range" oninput="setCols(this.value)">
            <span id="colLbl" class="col-label">6×</span>
          </div>
        </div>
        <div class="set-row">
          <label>Filter anzeigen</label>
          <div class="set-inline set-filter">
            <label class="set-check"><input type="checkbox" id="fpz-online" onchange="setFilterPill('online',this.checked)"> Online</label>
            <label class="set-check"><input type="checkbox" id="fpz-offline" onchange="setFilterPill('offline',this.checked)"> Offline</label>
            <label class="set-check"><input type="checkbox" id="fpz-paused" onchange="setFilterPill('paused',this.checked)"> Pausiert</label>
            <label class="set-check"><input type="checkbox" id="fpz-favs" onchange="setFilterPill('favs',this.checked)"> Favoriten</label>
            <label class="set-check"><input type="checkbox" id="fpz-fertig" onchange="setFilterPill('fertig',this.checked)"> Fertig</label>
          </div>
        </div>
      </div>
    </div>

    <div class="stcard">
      <h2>💡 Sonderbeleuchtung bei Fehler</h2>
      <p class="setup-info">Modelle ohne eigene Signalleuchte lassen bei einer Störung die Kammerbeleuchtung blinken.
        Läuft der Druck wieder, geht sie zurück auf normales Licht.</p>
      <div class="set-grid">
        <div class="set-row">
          <label>Blinken bei Störung</label>
          <div class="seg">
            <button class="seg-btn" id="blink-on"  onclick="setErrorBlink(true)">Ein</button>
            <button class="seg-btn" id="blink-off" onclick="setErrorBlink(false)">Aus</button>
          </div>
        </div>
        <div class="set-row">
          <label>Betroffene Modelle</label>
          <span id="blink-models" class="mono-note">–</span>
        </div>
      </div>
    </div>

    <div class="stcard">
      <h2>🏁 Startansicht</h2>
      <p class="setup-info" id="startfilter-desc">Welcher Filter beim Start der Übersicht aktiv ist.</p>
      <div class="set-grid">
        <div class="set-row">
          <label id="startfilter-label">Startfilter</label>
          <select id="start-filter" class="set-select" onchange="setStartFilter(this.value)">
            <option value="all">Alle</option>
            <option value="online">Nur Online</option>
            <option value="offline">Nur Offline</option>
          </select>
        </div>
      </div>
    </div>

    <div class="stcard">
      <h2 data-i18n="jpTitle">🧮 Job-Planer</h2>
      <p class="setup-info" data-i18n="jpDesc">Rechnet aus, wie viele Druckbretter du für eine Zielmenge brauchst und wie lange es dauert.</p>
      <div class="set-grid">
        <div class="set-row"><label data-i18n="jpPick">Aus Tabelle (Modell · Datei)</label>
          <select id="jp-pick" class="set-select" onchange="boardPick()"><option value="" data-i18n="jpPickNone">— manuell —</option></select></div>
        <div class="set-row"><label data-i18n="jpTarget">Zielmenge (Figuren)</label><input type="number" id="jp-target" class="set-select" min="1" value="200"></div>
        <div class="set-row"><label data-i18n="jpPerPlate">Figuren pro Brett</label><input type="number" id="jp-perplate" class="set-select" min="1" value="10"></div>
        <div class="set-row"><label data-i18n="jpTimePerPlate">Zeit pro Brett (Std.)</label><input type="number" id="jp-time" class="set-select" min="0" step="0.1" value="4"></div>
        <div class="set-row"><label data-i18n="jpPrinters">Drucker einsetzen</label>
          <select id="jp-printers" class="set-select">
            <option value="free" data-i18n="jpFree">Nur freie Drucker</option>
            <option value="all" data-i18n="jpAll">Alle Drucker</option>
          </select></div>
        <div class="set-row"><label></label><button class="hbtn primary" onclick="calcJobPlan()" data-i18n="jpCalc">Berechnen</button></div>
      </div>
      <div id="jp-result" class="jp-result mt8" style="display:none;"></div>

      <h3 class="mt8" data-i18n="btTitle">📋 Brett-Tabelle</h3>
      <p class="setup-info" data-i18n="btDesc">Pro Druckermodell und Druckdatei: Figuren pro Brett und Stunden pro Brett. Diese Werte füllt der Planer oben automatisch.</p>
      <table class="bt-table" id="bt-table">
        <thead><tr>
          <th data-i18n="btModel">Modell</th><th data-i18n="btFile">Druckdatei</th>
          <th data-i18n="btPerPlate">Figuren/Brett</th><th data-i18n="btHours">Std./Brett</th><th></th>
        </tr></thead>
        <tbody id="bt-body"></tbody>
      </table>
      <div class="mt8">
        <button class="hbtn btn-sm" onclick="boardTableAddRow()" data-i18n="btAdd">+ Zeile</button>
        <button class="hbtn btn-sm primary" onclick="boardTableSave()" data-i18n="btSave">Speichern</button>
      </div>
    </div>

    <div class="stcard">
      <h2 data-i18n="bcTitle">📡 Sendeseite (Stream)</h2>
      <p class="setup-info" data-i18n="bcDesc">Entkoppelte Nur-Ansicht zum Streamen (OBS) oder Veröffentlichen. Öffnet in einem eigenen Fenster, unabhängig von der Bedienung.</p>
      <div class="set-grid">
        <div class="set-row">
          <label data-i18n="bcView">Ansicht</label>
          <select id="bc-view" class="set-select" onchange="saveBroadcastCfg()">
            <option value="all" data-i18n="bcViewAll">Alle</option>
            <option value="online" data-i18n="bcViewOnline">Nur Online</option>
            <option value="offline" data-i18n="bcViewOffline">Nur Offline</option>
          </select>
        </div>
        <div class="set-row">
          <label data-i18n="bcCols">Spalten</label>
          <input type="number" id="bc-cols" class="set-select" min="1" max="20" value="9" onchange="saveBroadcastCfg()">
        </div>
        <div class="set-row">
          <label data-i18n="bcClick">Klicken erlauben</label>
          <input type="checkbox" id="bc-click" onchange="saveBroadcastCfg()">
        </div>
        <div class="set-row">
          <label data-i18n="bcUrl">Adresse</label>
          <div class="set-inline">
            <input type="text" id="bc-url" class="repo-input" readonly>
            <button class="hbtn" onclick="openBroadcast()" data-i18n="bcOpen">Öffnen</button>
          </div>
        </div>
      </div>
    </div>

    <div class="stcard">
      <h2 data-i18n="cfTitle">🌐 Cloudflare-Tunnel</h2>
      <p class="setup-info" data-i18n="cfDesc">Veröffentlicht die Sendeseite im Internet — ohne Portfreigabe, IP versteckt. Nur die Sendeseite ist erreichbar, keine Steuerung.</p>
      <div id="cf-need" class="setup-info" style="display:none;">
        <span data-i18n="cfNeed">cloudflared wird benötigt (einmaliger Download).</span>
        <button class="hbtn" onclick="installTunnel()" data-i18n="cfInstall">Installieren</button>
        <span id="cf-inst-status" class="note"></span>
      </div>
      <div class="set-grid">
        <div class="set-row">
          <label data-i18n="cfMode">Zugang</label>
          <select id="cf-mode" class="set-select" onchange="saveTunnelCfg()">
            <option value="public" data-i18n="cfPublic">Öffentlich</option>
            <option value="secret" data-i18n="cfSecret">Geheim (Token)</option>
          </select>
        </div>
        <div class="set-row">
          <label data-i18n="cfClick">Klicken erlauben</label>
          <input type="checkbox" id="cf-click" onchange="saveTunnelCfg()">
        </div>
        <div class="set-row">
          <label data-i18n="cfStatus">Status</label>
          <span id="cf-status" class="mono-value">–</span>
        </div>
        <div class="set-row">
          <label data-i18n="cfUrl">Adresse</label>
          <div class="set-inline">
            <input type="text" id="cf-url" class="repo-input" readonly>
            <button class="hbtn" onclick="copyTunnelUrl()" data-i18n="cfCopy">Kopieren</button>
          </div>
        </div>
        <div class="set-row">
          <label></label>
          <div class="set-inline">
            <button class="hbtn primary" id="cf-start" onclick="startTunnelUI()" data-i18n="cfStart">Veröffentlichen</button>
            <button class="hbtn" id="cf-stop" onclick="stopTunnelUI()" data-i18n="cfStop">Stoppen</button>
          </div>
        </div>
      </div>
    </div>

    <div class="stcard">
      <h2 data-i18n="ytTitle">📺 YouTube-Livestream</h2>
      <p class="setup-info" data-i18n="ytDesc">Schiebt ein Kachel-Mosaik der Kameras direkt an YouTube. Stream-Key aus YouTube Studio → Livestreaming. Standardwerte sind gesetzt, lassen sich aber anpassen.</p>
      <div class="set-grid">
        <div class="set-row">
          <label data-i18n="ytKey">Stream-Key</label>
          <input type="password" id="yt-key" class="repo-input" placeholder="xxxx-xxxx-xxxx-xxxx" onchange="saveYouTubeCfg()">
        </div>
        <div class="set-row">
          <label data-i18n="ytRes">Auflösung</label>
          <input type="text" id="yt-res" class="set-select" placeholder="1920x1080" onchange="saveYouTubeCfg()">
        </div>
        <div class="set-row">
          <label data-i18n="ytBitrate">Bitrate (kbit/s)</label>
          <input type="number" id="yt-bitrate" class="set-select" min="500" max="51000" step="500" value="4500" onchange="saveYouTubeCfg()">
        </div>
        <div class="set-row">
          <label data-i18n="ytFps">Bilder/s</label>
          <input type="number" id="yt-fps" class="set-select" min="1" max="60" value="15" onchange="saveYouTubeCfg()">
        </div>
        <div class="set-row">
          <label data-i18n="ytStatus">Status</label>
          <span id="yt-status" class="mono-value">–</span>
        </div>
        <div class="set-row">
          <label></label>
          <div class="set-inline">
            <button class="hbtn primary" id="yt-start" onclick="startYouTubeUI()" data-i18n="ytStart">Stream starten</button>
            <button class="hbtn" id="yt-stop" onclick="stopYouTubeUI()" data-i18n="ytStop">Stoppen</button>
          </div>
        </div>
      </div>
    </div>

    <div class="stcard">
      <h2 data-i18n="aiTitle">🤖 KI-Assistent</h2>
      <p class="setup-info" data-i18n="aiDesc">Bindet OpenAI (ChatGPT) oder Anthropic (Claude) mit deinem eigenen API-Key an — für Fehler-Erklärung und Kamera-Fehlererkennung. Der Key bleibt lokal, wird nie geloggt oder nach außen gegeben.</p>
      <div class="set-grid">
        <div class="set-row"><label data-i18n="aiConnectors">KI-Konnektoren</label>
          <div class="ai-conn-wrap">
            <div id="ai-conn-list" class="ai-conn-list"></div>
            <div class="ai-conn-add">
              <input type="text" id="ai-new-name" class="repo-input" placeholder="Name (z. B. GPT-4o, lokal…)">
              <select id="ai-new-provider" class="set-select">
                <option value="openai">OpenAI (ChatGPT)</option>
                <option value="anthropic">Anthropic (Claude)</option>
                <option value="ollama" data-i18n="aiOllama">Ollama (lokal)</option>
              </select>
              <button class="hbtn primary" onclick="aiAddConnector()" data-i18n="aiAddConn">+ Konnektor</button>
            </div>
          </div>
        </div>
        <div class="set-row" id="ai-conn-edit-row" style="display:none"><label data-i18n="aiEditConn">Konnektor bearbeiten</label>
          <div class="ai-conn-edit">
            <input type="hidden" id="ai-edit-id">
            <div class="ai-edit-f"><label data-i18n="aiConnName">Name</label><input type="text" id="ai-edit-name" class="repo-input"></div>
            <div class="ai-edit-f"><label data-i18n="aiProvider">Anbieter</label>
              <select id="ai-edit-provider" class="set-select" onchange="aiEditProviderChange()">
                <option value="openai">OpenAI (ChatGPT)</option>
                <option value="anthropic">Anthropic (Claude)</option>
                <option value="ollama" data-i18n="aiOllama">Ollama (lokal)</option>
              </select></div>
            <div class="ai-edit-f"><label data-i18n="aiModel">Modell (optional)</label><input type="text" id="ai-edit-model" class="repo-input" placeholder="gpt-4o-mini / claude-3-5-sonnet-latest / llama3.2-vision"></div>
            <div class="ai-edit-f" id="ai-edit-key-row"><label data-i18n="aiKey">API-Key</label><input type="password" id="ai-edit-key" class="repo-input" placeholder="sk-…"></div>
            <div class="ai-edit-f" id="ai-edit-url-row"><label data-i18n="aiOllamaUrl">Ollama-Adresse</label><input type="text" id="ai-edit-url" class="repo-input" placeholder="http://localhost:11434"></div>
            <div class="set-inline">
              <button class="hbtn primary" onclick="aiSaveConnector()" data-i18n="save">Speichern</button>
              <button class="hbtn" onclick="aiCancelEdit()" data-i18n="cancel">Abbrechen</button>
            </div>
          </div>
        </div>
        <div class="set-row"><label data-i18n="aiStatus">Status</label><span id="ai-status" class="mono-value">–</span></div>
        <div class="set-row"><label data-i18n="aiPromptExplain">Prompt: Fehler erklären</label><textarea id="ai-prompt-explain" class="repo-input ai-ta" rows="3" onchange="saveAIPrompts()"></textarea></div>
        <div class="set-row"><label data-i18n="aiPromptVision">Prompt: Bild prüfen</label><textarea id="ai-prompt-vision" class="repo-input ai-ta" rows="3" onchange="saveAIPrompts()"></textarea></div>
        <div class="set-row"><label></label><span class="note-sm" data-i18n="aiPromptHint">Platzhalter: {lang} = Sprache, {error} = Fehlermeldung</span></div>
      </div>
    </div>

    <div class="stcard admin-only" id="admin-access-card">
      <h2 data-i18n="accessTitle">👥 Benutzer & Zugang</h2>
      <p class="setup-info" data-i18n="accessDesc">Einzelne Konten mit Rollen: Viewer (nur ansehen), Operator (ansehen + steuern), Admin (alles). Programme greifen über API-Tokens zu. Passwörter und Zugangscodes bleiben lokal.</p>

      <h3 class="set-sub" data-i18n="accessUsers">Benutzer</h3>
      <div id="users-list" class="acc-list"></div>
      <div class="acc-add">
        <input type="text" id="new-user-name" class="repo-input" placeholder="Benutzername">
        <input type="password" id="new-user-pass" class="repo-input" placeholder="Passwort (min. 6)">
        <select id="new-user-role" class="set-select">
          <option value="viewer" data-i18n="roleViewer">Viewer</option>
          <option value="operator" data-i18n="roleOperator">Operator</option>
          <option value="admin" data-i18n="roleAdmin">Admin</option>
        </select>
        <button class="hbtn primary" onclick="addUser()" data-i18n="accessAddUser">+ Benutzer</button>
      </div>

      <h3 class="set-sub" data-i18n="accessTokens">API-Tokens</h3>
      <p class="note-sm" data-i18n="accessTokensHint">Für andere Programme. Das Token wird nur einmal beim Erstellen angezeigt — danach nicht mehr abrufbar.</p>
      <div id="tokens-list" class="acc-list"></div>
      <div class="acc-add">
        <input type="text" id="new-token-name" class="repo-input" placeholder="Token-Name (z. B. ERP)">
        <select id="new-token-role" class="set-select">
          <option value="viewer" data-i18n="roleViewer">Viewer</option>
          <option value="operator" data-i18n="roleOperator">Operator</option>
          <option value="admin" data-i18n="roleAdmin">Admin</option>
        </select>
        <button class="hbtn primary" onclick="addToken()" data-i18n="accessAddToken">+ Token</button>
      </div>
      <div id="token-reveal" class="token-reveal" style="display:none;"></div>
    </div>

    <div class="stcard">
      <h2>⬆ Programmversion</h2>
      <div class="set-grid">
        <div class="set-row">
          <label>Installiert</label>
          <span id="upd-current" class="mono-value">–</span>
        </div>
        <div class="set-row">
          <label>GitHub-Adresse</label>
          <div class="set-inline">
            <input type="text" id="upd-repo" class="repo-input" placeholder="https://github.com/besitzer/name">
            <button class="hbtn" onclick="saveUpdateRepo()">Speichern</button>
          </div>
        </div>
        <div class="set-row">
          <label>Lokaler Speicherort</label>
          <span id="upd-datadir" class="mono-note breakable">–</span>
        </div>
      </div>
      <div class="btn-row">
        <button class="hbtn" onclick="checkUpdate(true)" id="btn-updcheck">Nach Update suchen</button>
        <button class="hbtn green hidden" onclick="installUpdate()" id="btn-updinstall">Update einspielen</button>
        <span id="upd-status" class="muted-note"></span>
      </div>
      <div id="upd-progress" class="progress-wrap hidden">
        <div class="progress-track"><div id="upd-bar" class="progress-fill"></div></div>
        <p id="upd-progress-text" class="progress-text"></p>
      </div>
      <p id="upd-notes" class="release-notes"></p>
    </div>

    <div class="stcard copyright-card">
      <p>© Brickshouse GmbH · Druckerfarm.ch</p>
      <p class="oss-hinweis">Nutzt quelloffene Software: paho.mqtt.golang (EPL-2.0/EDL-1.0), gorilla/websocket (BSD-2), jlaffaye/ftp (ISC), hashicorp/go-multierror &amp; errwrap (MPL-2.0), golang.org/x/net &amp; x/sync (BSD-3). Zur Laufzeit: go2rtc (MIT), ffmpeg (BtbN, GPL). Einzelheiten in <a href="#" onclick="openLicenses();return false;" class="oss-link">THIRD_PARTY_LICENSES.md</a>.</p>
    </div>
  </div>
</div>

<div class="view hidden" id="view-jobs">
  <div class="jobs-page">
    <div class="jobs-col jobs-col-run">
      <h2 data-i18n="jobsRunning">⚙ Laufende Aufträge</h2>
      <div id="jobs-list" class="jobs-list"></div>
    </div>
    <div class="jobs-col jobs-col-plan">
      <div class="plan-head">
        <h2 data-i18n="planTitle">🗺 Job-Planer</h2>
        <div class="plan-head-btns">
          <button class="hbtn" onclick="loadPlan()" data-i18n="loadData">↺ Daten laden</button>
          <button class="hbtn primary" onclick="savePlan()" data-i18n="save">Speichern</button>
        </div>
      </div>
      <p class="setup-info" data-i18n="planDesc">Plane pro Design-Figur, wie viele Stück auf welchem Drucker-Modell gedruckt werden sollen. Spalten = vorhandene Modelle, Zeilen = Figuren.</p>
      <div class="plan-add">
        <input type="text" id="plan-new-name" class="repo-input" placeholder="Neue Figur (Name)…">
        <button class="hbtn primary" onclick="planAddFigure()" data-i18n="planAddFig">+ Figur</button>
      </div>
      <div id="plan-matrix" class="plan-matrix-wrap"></div>
    </div>
  </div>
</div>
<div class="view hidden" id="view-fsv"></div>
<div class="view hidden" id="view-fsc"></div>

<!-- SYNC VIEW -->
<div class="view" id="view-sync">
  <div class="sp">
    <!-- Upload & Push -->
    <div class="stcard">
      <h2>🗂 Druckdateien verwalten</h2>

      <!-- Dauerhafter Hinweis: der Dateizugriff braucht LAN-/Entwicklermodus + Zugangscode; nur Port 990. -->
      <div class="fs-hint" id="fs-perm-hint" data-i18n="fsPermHint">Dateizugriff läuft ausschließlich über den Drucker-Speicher auf Port 990. Antwortet ein Drucker nicht: am Gerät den LAN-/Entwicklermodus einschalten und den Zugangscode hinterlegen — einen anderen Port bietet der Drucker nicht.</div>

      <!-- Farmweiter FTP-/Speicher-Check -->
      <div class="row-10">
        <button class="hbtn btn-xs" onclick="runFtpCheck()" data-i18n="ftpCheckBtn">FTP-/Speicher-Check (alle Drucker)</button>
        <span class="note-sm" id="ftpcheck-sum"></span>
      </div>
      <div id="ftpcheck-result" class="mt8"></div>

      <!-- Dateisuche ueber mehrere Drucker -->
      <div class="stcard card-inset">
        <div class="row-10">
          <input type="text" id="fileSearch" placeholder="Dateiname suchen (ab 3 Zeichen)"
                 oninput="onSearchTyping()" onkeydown="if(event.key==='Enter')toggleSearch()"
                 class="search-field">
          <button class="hbtn" id="btn-search" onclick="toggleSearch()">Suchen</button>
          <span id="fileSearchStatus" class="note"></span>
        </div>
        <div class="bar-track-slim hidden" id="searchBarWrap">
          <div id="searchBar" class="bar-fill"></div>
        </div>
        <div id="searchSelBar" class="sel-bar hidden">
          <span id="searchSelCount" class="fs12-text"></span>
          <button class="hbtn" onclick="selectAllHits(true)">Alle auswählen</button>
          <button class="hbtn" onclick="selectAllHits(false)">Auswahl aufheben</button>
          <button class="hbtn btn-danger" onclick="deleteSelectedHits()">Ausgewählte löschen</button>
          <span id="searchDelStatus" class="note"></span>
        </div>
      </div>

      <!-- Model filter -->
      <div class="sync-filter-bar mt10" id="sync-model-filter"></div>

      <!-- Modus-Wahl -->
      <div class="stcard card-inset mt10" id="sync-mode-card">
        <div class="row-10" style="gap:18px;flex-wrap:wrap;">
          <label class="cbx"><input type="radio" name="sync-mode" value="manual" onchange="setSyncMode('manual')"> <span data-i18n="syncModeManual">Manueller Upload</span></label>
          <label class="cbx"><input type="radio" name="sync-mode" value="auto" onchange="setSyncMode('auto')"> <span data-i18n="syncModeAuto">Automatischer Ordner-Sync</span></label>
          <label class="cbx"><input type="radio" name="sync-mode" value="both" onchange="setSyncMode('both')"> <span data-i18n="syncModeBoth">Beides</span></label>
        </div>
        <div id="sync-auto-cfg" class="set-grid mt8" style="display:none;">
          <div class="set-row"><label data-i18n="syncFolder">Lokaler Ordner</label><input type="text" id="sync-ref-path" class="repo-input" placeholder="C:\\Drucke\\Freigabe"></div>
          <div class="set-row"><label data-i18n="syncSdPath">Ordner auf SD</label><input type="text" id="sync-sd-path" class="repo-input" placeholder="/"></div>
          <div class="set-row"><label data-i18n="syncInterval">Intervall (Minuten)</label><input type="number" id="sync-interval" class="set-select" min="1" max="1440" value="60"></div>
          <div class="set-row"><label><input type="checkbox" id="sync-auto"> <span data-i18n="syncAutoOn">Automatisch abgleichen</span></label></div>
          <div class="set-row"><label data-i18n="syncModelFolders">Ordner je Modell</label><div id="sync-model-folders" style="flex:1;min-width:0;"></div></div>
          <div class="set-row"><label></label><button class="hbtn" onclick="checkSyncDiff()" data-i18n="syncCheckDiff">Abweichungen prüfen</button></div>
          <div class="set-row"><label></label><button class="hbtn" onclick="saveSyncConfig()" data-i18n="save">Speichern</button></div>
        </div>
        <div id="sync-diff-hint" class="sync-diff-hint mt8" style="display:none;"></div>
      </div>

      <!-- Drop zone -->
      <div class="drop-zone" id="syncDropZone" onclick="document.getElementById('syncFileInput').click()">
        <div id="syncDropText">📂 Dateien hier reinziehen oder klicken<br><span class="hint-xs">.gcode / .3mf Dateien</span></div>
        <div class="drop-file-list" id="syncDropList"></div>
      </div>
      <input type="file" id="syncFileInput" multiple accept=".gcode,.3mf" class="hidden">

      <!-- Printer selection -->
      <div class="row-10-mb">
        <button class="hbtn btn-xs" id="sync-toggle-btn" onclick="syncToggleAll()">☑ Alle abwählen</button>
        <span class="note-sm" id="sync-sel-count">0 Drucker ausgewählt</span>
        <button class="hbtn btn-xs push-right" onclick="loadAllSDFiles()" id="btn-loadall">↺ Inhalte laden</button>
      </div>
      <div class="sync-printer-list" id="syncPrinterList"></div>

      <!-- Medien / Speicher: über die oben ausgewählten Drucker anwendbar -->
      <div class="stcard card-inset mt10" id="sync-media-card">
        <div class="row-10">
          <strong data-i18n="mediaFarmTitle">Medien / Speicher (Auswahl)</strong>
          <span class="note-sm" id="sync-media-note" data-i18n="mediaFarmNote">Wirkt auf die oben ausgewählten Drucker.</span>
        </div>
        <div class="pc-row mt8">
          <button class="pc-btn danger" onclick="farmMediaClear('video')" data-i18n="mediaClearVideo">Alle Videos löschen</button>
          <button class="pc-btn danger" onclick="farmMediaClear('snapshot')" data-i18n="mediaClearSnap">Alle Schnappschüsse löschen</button>
          <button class="pc-btn" onclick="farmTimelapse(false,true)" data-i18n="mediaTlOffAll">Zeitraffer aus + dauerhaft</button>
          <button class="pc-btn" onclick="farmTimelapse(true,false)" data-i18n="mediaTlOnAll">Zeitraffer an</button>
        </div>
        <div id="sync-media-result" class="mt8"></div>
      </div>

      <!-- Upload button + progress -->
      <div class="row-8-mt">
        <button class="sync-upload-btn" id="syncUploadBtn" onclick="startDropSync()" disabled>⬆ Hochladen</button>
        <button class="hbtn btn-xs hidden" id="syncPauseBtn" onclick="toggleUploadPause()">⏸ Pause</button>
        <button class="hbtn btn-xs btn-danger hidden" id="syncCancelBtn" onclick="cancelUpload()">✕ Abbrechen</button>
        <span id="syncUploadStatus" class="note-sm"></span>
      </div>
      <div id="syncUploadProgress" class="mt8 hidden"></div>

      <!-- Historie: bleibt ueber Neustarts hinweg erhalten -->
      <div class="stcard card-inset mt10">
        <div class="row-10">
          <button class="hbtn btn-xs" onclick="loadUploadLog()">🕘 Historie laden</button>
          <span class="note-sm" id="ul-zusammenfassung"></span>
          <button class="hbtn btn-xs btn-danger push-right" id="btn-ul-leeren" onclick="leereUploadLog()" disabled>Historie leeren</button>
        </div>
        <div id="ul-liste" class="ul-liste"></div>
      </div>
    </div>
  </div>
</div>
</main>

<div class="toasts" id="toasts"></div>

<div id="dd-backdrop" onclick="closeDD()"></div>
<div id="dd-pop"></div>

<!-- Setup-Dialog: erscheint nur, wenn go2rtc oder ffmpeg fehlen -->
<div id="setup-overlay" class="modal-backdrop hidden">
  <div class="modal-box">
    <h2 class="modal-title">Komponenten fehlen</h2>
    <p class="modal-lead">
      Drucker, Status und Datei-Sync funktionieren bereits. Für Video und Snapshots fehlen noch:
    </p>
    <div id="setup-list" class="modal-list"></div>
    <div id="setup-progress" class="modal-progress hidden">
      <div class="bar-track">
        <div id="setup-bar" class="bar-fill"></div>
      </div>
      <p id="setup-progress-text" class="progress-text"></p>
    </div>
    <p id="setup-error" class="modal-error hidden"></p>
    <div class="modal-actions">
      <button class="hbtn" onclick="dismissSetup(true)"  id="setup-never">Nicht mehr fragen</button>
      <button class="hbtn" onclick="dismissSetup(false)" id="setup-later">Später</button>
      <button class="hbtn green" onclick="startSetupDownload()" id="setup-go">Jetzt herunterladen</button>
    </div>
    <p class="modal-footnote">
      Quellen: go2rtc von github.com/AlexxIT/go2rtc, ffmpeg von github.com/BtbN/FFmpeg-Builds.
      Beide werden nach dem Download per SHA256 geprüft und landen im Programmordner.
    </p>
  </div>
</div>

<script>
// ─── I18N ─────────────────────────────────────────────────────────────────────
const LANGS = __LANGS_JSON__;
const OVERLAYS = __OVERLAYS_JSON__;
let UEB = {};
window.BROADCAST_OPTS = __BROADCAST_OPTS__;

let LANG = 'de';
const APP_START = Date.now();
let G2_PREV = null;
let G2_FAILS = 0;
let THEME = 'light';
let ERROR_BLINK = true;
let START_FILTER = "all";
let BLINK_MODELS = [];
let BLINK_AUS = {};   // ip -> true, wenn Blinken für diesen Drucker aus
let KAMERA_AUS = {};  // ip -> true, wenn Kamera dauerhaft aus („Private")
let MEDIA_ENFORCE = {}; // ip -> true, wenn Zeitraffer dauerhaft aus gehalten wird
function isCameraOff(ip){ return !!KAMERA_AUS[ip]; }
function isBlinkOff(ip){ return !!BLINK_AUS[ip]; }
function isBlinkModel(model){ const m=String(model||"").toUpperCase(); return (BLINK_MODELS||[]).some(x=>m===String(x).toUpperCase()); }
function t(k) { return LANGS[LANG][k] || LANGS['de'][k] || k; }

// ─── AUTOMATIC ENGLISH TRANSLATION ────────────────────────────────────────────
//
// German is the source language in the code. In English a
// DOM translator translates the rendered text via the DEEN table — including
// dynamically generated text, because a MutationObserver translates every change. So
// not every single text spot in the code has to be touched.
let UEB_OBS = null;

function uebText(node) {
  const p = node.parentNode;
  if (p && (p.nodeName === 'SCRIPT' || p.nodeName === 'STYLE')) return;
  const raw = node.nodeValue;
  if (!raw) return;
  const norm = raw.replace(/\s+/g, ' ').trim();
  if (norm && UEB[norm] !== undefined) {
    const lead = (raw.match(/^\s*/) || [''])[0];
    const trail = (raw.match(/\s*$/) || [''])[0];
    node.nodeValue = lead + UEB[norm] + trail;
  }
}

function uebAttr(el) {
  if (!el.getAttribute) return;
  ['title', 'placeholder'].forEach(a => {
    if (el.hasAttribute(a)) {
      const v = (el.getAttribute(a) || '').replace(/\s+/g, ' ').trim();
      if (v && UEB[v] !== undefined) el.setAttribute(a, UEB[v]);
    }
  });
}

function translateTree(root) {
  if (!UEB || !Object.keys(UEB).length || !root) return;
  if (root.nodeType === 1) uebAttr(root);
  if (root.querySelectorAll) root.querySelectorAll('[title],[placeholder]').forEach(uebAttr);
  const start = (root.nodeType === 3) ? null : root;
  if (root.nodeType === 3) { uebText(root); return; }
  const w = document.createTreeWalker(start, NodeFilter.SHOW_TEXT, null);
  const nodes = []; let n;
  while ((n = w.nextNode())) nodes.push(n);
  nodes.forEach(uebText);
}

function startTranslator() {
  if (!UEB || !Object.keys(UEB).length) return;
  translateTree(document.body);
  if (UEB_OBS) return;
  UEB_OBS = new MutationObserver(muts => {
    for (const m of muts) {
      m.addedNodes && m.addedNodes.forEach(nd => {
        if (nd.nodeType === 3) uebText(nd);
        else if (nd.nodeType === 1) translateTree(nd);
      });
    }
  });
  UEB_OBS.observe(document.body, { childList: true, subtree: true });
}

// enableTranslation picks the overlay table for the current language and
// starts the live translator. For German (no overlay) nothing happens.
function enableTranslation() {
  UEB = OVERLAYS[LANG] || {};
  if (Object.keys(UEB).length) startTranslator();
}

function setLang(l) {
  // Switching reloads the page: German is the source, English is translated
  // automatically on load. So the switch is clean in both directions.
  markSeg('lang-', l);
  fetch('/api/settings', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({lang: l})})
    .catch(()=>{})
    .finally(() => location.reload());
}

// showLanguagePicker shows the language choice on the very first start.
function showLanguagePicker() {
  const el = document.getElementById('lang-picker');
  if (el) el.style.display = 'flex';
}

// pickLanguage applies the choice, remembers it permanently and closes the
// overlay. After that the prompt does not return.
function openLicenses() {
  fetch('/api/open-licenses', {method:'POST'})
    .then(() => toast('THIRD_PARTY_LICENSES.md im Explorer geöffnet', 'ok'))
    .catch(() => toast('Konnte die Datei nicht öffnen', 'er'));
}

function pickLanguage(l) {
  LANG = ['de','en','zh','es'].includes(l) ? l : 'de';
  markSeg('lang-', LANG);
  applyLang();
  render();
  saveSettings({lang: LANG, sprache_gewaehlt: true});
  const el = document.getElementById('lang-picker');
  if (el) el.style.display = 'none';
  enableTranslation();
}


// ─── CUSTOM COLORS ────────────────────────────────────────────────────────────
// Color customization is gone — there are only the two schemes light
// and dark. Previously stored values are cleared once, otherwise
// they would keep overriding the chosen scheme.
function purgeLegacyColors() {
  ['bg','text','btnbg','btntext','seg','menubg','menutext'].forEach(k => localStorage.removeItem('custom-' + k));
  ['--bg','--s1','--s2','--s3','--text','--muted','--btn-bg','--btn-text','--seg-bg','--menu-bg','--menu-text']
    .forEach(v => document.documentElement.style.removeProperty(v));
}

function applyLang() {
  (el=document.getElementById('tab-grid')) && (el.textContent = t('tabGrid'));
  (el=document.getElementById('tab-list')) && (el.textContent = t('tabList'));
  (el=document.getElementById('tab-setup-lbl')) && (el.textContent = t('tabSetup'));
  (el=document.getElementById('tab-import-lbl')) && (el.textContent = t('tabImport'));
  (el=document.getElementById('tab-import-lbl2')) && (el.textContent = t('tabImport'));
  (el=document.getElementById('lbl-filter')) && (el.textContent = t('filter'));
  (el=document.getElementById('lbl-total')) && (el.textContent = t('total'));
  (el=document.getElementById('lbl-online-pill')) && (el.textContent = t('online'));
  document.getElementById('searchInput').placeholder = t('search');
  (el=document.getElementById('btn-startall')) && (el.textContent = t('startAll'));
  (el=document.getElementById('btn-stopall')) && (el.textContent = t('stopAll'));
  (el=document.getElementById('imp-csv-title')) && (el.textContent = t('impCsvTitle'));
  (el=document.getElementById('imp-csv-desc')) && (el.textContent = t('impCsvDesc'));
  document.getElementById('dz-text').innerHTML = t('dzText');
  (el=document.getElementById('btn-clear')) && (el.textContent = t('btnClear'));
  (el=document.getElementById('imp-manual-title')) && (el.textContent = t('impManualTitle'));
  (el=document.getElementById('lbl-model')) && (el.textContent = t('lblModel'));
  (el=document.getElementById('lbl-name')) && (el.textContent = t('lblName'));
  (el=document.getElementById('lbl-ip')) && (el.textContent = t('lblIp'));
  (el=document.getElementById('lbl-code')) && (el.textContent = t('lblCode'));
  (el=document.getElementById('btn-add')) && (el.textContent = t('btnAdd'));
  (el=document.getElementById('lbl-serial')) && (el.textContent = t('lblSerial'));
  (el=document.getElementById('imp-saved-title')) && (el.textContent = t('impSavedTitle'));
  (el=document.getElementById('btn-delall')) && (el.textContent = t('btnDelAll'));
  (el=document.getElementById('cam-admin-title')) && (el.textContent = t('camAdminTitle'));
  (el=document.getElementById('cam-admin-desc')) && (el.textContent = t('camAdminDesc'));
  (el=document.getElementById('cam-name')) && (el.placeholder = t('camNamePh'));
  (el=document.getElementById('cam-src')) && (el.placeholder = t('camSourcePh'));
  (el=document.getElementById('cam-add-btn')) && (el.textContent = t('camAddBtn'));
  (el=document.getElementById('cam-hint')) && (el.textContent = t('camHint'));
  (el=document.getElementById('xiaomi-sum')) && (el.textContent = t('xmTitle'));
  (el=document.getElementById('xiaomi-desc')) && (el.textContent = t('xmDesc'));
  (el=document.getElementById('xm-user')) && (el.placeholder = t('xmUserPh'));
  (el=document.getElementById('xm-pass')) && (el.placeholder = t('xmPassPh'));
  (el=document.getElementById('xm-server')) && (el.placeholder = t('xmServerPh'));
  (el=document.getElementById('xm-btn')) && (el.textContent = t('xmBtn'));
  if(typeof renderCamAdmin==='function') renderCamAdmin();
  (el=document.getElementById('imp-export-title')) && (el.textContent = t('impExportTitle'));
  (el=document.getElementById('btn-expcsv')) && (el.textContent = t('btnExpCsv'));
  (el=document.getElementById('btn-expyaml')) && (el.textContent = t('btnExpYaml'));
  document.getElementById('setup-info').innerHTML = t('setupInfo');
  (el=document.getElementById('btn-check')) && (el.textContent = t('btnCheck'));
  (el=document.getElementById('tab-sync-lbl')) && (el.textContent = t('tabSync'));
  (el=document.getElementById('tab-sync-lbl2')) && (el.textContent = t('tabSync'));
  (el=document.getElementById('sync-title-cfg')) && (el.textContent = t('syncTitleCfg'));
  (el=document.getElementById('sync-title-ctrl')) && (el.textContent = t('syncTitleCtrl'));
  (el=document.getElementById('sync-title-log')) && (el.textContent = t('syncTitleLog'));
  (el=document.getElementById('sync-lbl-path')) && (el.textContent = t('syncLblPath'));
  (el=document.getElementById('sync-lbl-sdpath')) && (el.textContent = t('syncLblSdPath'));
  (el=document.getElementById('sync-lbl-auto')) && (el.textContent = t('syncLblAuto'));
  (el=document.getElementById('sync-lbl-interval')) && (el.textContent = t('syncLblInterval'));
  (el=document.getElementById('sync-btn-save')) && (el.textContent = t('syncBtnSave'));
  const _ssd = document.getElementById('sync-title-sd'); if(_ssd) _ssd.textContent = t('syncTitleSd');
  (el=document.getElementById('btn-restart')) && (el.textContent = t('btnRestart'));
  (el=document.getElementById('stitle')) && (el.textContent = t('statusChecking'));
  (el=document.getElementById('sdesc')) && (el.textContent = t('statusConnecting'));
  // Generic way: every element with data-i18n gets its text from the
  // dictionary. So new text can be translated without adding a line
  // here — key into the JSON, data-i18n on the element, done.
  document.querySelectorAll('[data-i18n]').forEach(el => {
    const v = t(el.getAttribute('data-i18n')); if (v) el.textContent = v;
  });
  document.querySelectorAll('[data-i18n-html]').forEach(el => {
    const v = t(el.getAttribute('data-i18n-html')); if (v) el.innerHTML = v;
  });
  document.querySelectorAll('[data-i18n-ph]').forEach(el => {
    const v = t(el.getAttribute('data-i18n-ph')); if (v) el.placeholder = v;
  });
  renderYaml();
}


// ─── WEBRTC MANAGER ───────────────────────────────────────────────────────────
const PCS = {};          // active RTCPeerConnection per IP
const STREAMS = {};      // live MediaStream per IP — ueberlebt einen View-Wechsel
const PAUSED = {};       // pause state per IP
const RETRY_TIMERS = {}; // pending reconnect timer per IP
const VERBINDE_TIMER = {}; // Zeitlimit: dreht der Ladekreis zu lange, wird Text gezeigt

// attachVideo attaches an already running stream to the <video> of the currently
// visible view. This way switching to fullscreen needs no
// neue Verbindung — frueher baute jede Vollbildansicht ihre eigenen
// PeerConnections and thus doubled the load.
function attachVideo(ip) {
  const v = document.getElementById('vid-' + ip);
  if (!v || !STREAMS[ip]) return false;
  if (v.srcObject !== STREAMS[ip]) {
    v.srcObject = STREAMS[ip];
    // As soon as an image actually arrives, remove the "connecting …" overlay.
    const fertig = () => { const o = document.getElementById('conn-' + ip); if (o) o.remove(); hideFsSpin(ip);
      if (VERBINDE_TIMER[ip]) { clearTimeout(VERBINDE_TIMER[ip]); delete VERBINDE_TIMER[ip]; } };
    v.onplaying = fertig; v.onloadeddata = fertig;
    v.play().catch(() => {});
  }
  return true;
}

// showNoImage replaces a tile's spinner with text when no image
// arrives. Whether "Offline" or "No image" depends on whether the printer is
// still reachable via MQTT.
function showNoImage(ip) {
  const conn = document.getElementById('conn-' + ip);
  if (!conn) return;
  const s = STATUS[ip];
  const txt = isCamera(ip) ? t('camUnavailable') : ((s && s.online) ? t('camNoImage') : t('lblOffline'));
  conn.innerHTML = '<div class="cip">' + txt + '</div>';
  conn.classList.add('cam-kein-bild');
}

// mountStreams supplies freshly rendered tiles: existing streams are
// reattached, missing ones built up staggered.
function mountStreams(printers) {
  if (SNAP_MODE > 0) { startSnapAll(printers); return; }
  let started = 0;
  printers.forEach(p => {
    if (ST[p.ip] !== 'live') return;
    // Offline or in repair: no connection, no spinner.
    const stx = STATUS[p.ip];
    if (!(stx && stx.online) || isRepair(p.ip) || isCameraOff(p.ip)) return;
    if (PCS[p.ip] && attachVideo(p.ip)) return;
    const delay = 100 + (started++) * 80;
    setTimeout(() => { if (ST[p.ip] === 'live') startWebRTC(p.ip, sname(p)); }, delay);
  });
}

// Tear a PeerConnection down completely. Every path that abandons a pc has to
// go through this — pc.close() alone leaves the keepalive interval and the ICE
// timeout running, and simply dropping the reference (as the old error handler
// did) leaves the connection and its video decoder alive for good.
function killPC(pc) {
  if (!pc) return;
  try {
    if (pc._keepAliveTimer) { clearInterval(pc._keepAliveTimer); pc._keepAliveTimer = null; }
    if (pc._iceTimeout)     { clearTimeout(pc._iceTimeout);      pc._iceTimeout = null; }
    if (pc._iceCheck)       { pc.removeEventListener('icegatheringstatechange', pc._iceCheck); pc._iceCheck = null; }
    pc.ontrack = null;
    pc.onconnectionstatechange = null;
    pc.onicecandidate = null;
    if (pc._dc) {
      pc._dc.onopen = null; pc._dc.onclose = null;
      try { pc._dc.close(); } catch(e) {}
      pc._dc = null;
    }
    try { pc.getReceivers().forEach(r => { if (r.track) r.track.stop(); }); } catch(e) {}
    pc.close();
  } catch(e) {}
}

// At most one pending reconnect per printer. Without this guard every closed
// connection scheduled its own retry, and the retries multiplied each other.
function scheduleRetry(ip, streamName, delay) {
  if (RETRY_TIMERS[ip]) return;
  RETRY_TIMERS[ip] = setTimeout(() => {
    delete RETRY_TIMERS[ip];
    if (ST[ip] === 'live' && G2online) startWebRTC(ip, streamName);
  }, delay);
}

async function startWebRTC(ip, streamName) {
  stopWebRTC(ip);

  // No STUN — everything is local LAN, no external servers needed
  const pc = new RTCPeerConnection({
    iceServers: [],
    iceTransportPolicy: 'all',
    bundlePolicy: 'max-bundle',
    rtcpMuxPolicy: 'require'
  });
  PCS[ip] = pc;
  PAUSED[ip] = false;

  // Time limit for the spinner: if no image arrives within 10 s (printer off,
  // camera unreachable), the spinner disappears and text is shown.
  if (VERBINDE_TIMER[ip]) clearTimeout(VERBINDE_TIMER[ip]);
  VERBINDE_TIMER[ip] = setTimeout(() => {
    if (PCS[ip] === pc && !STREAMS[ip]) showNoImage(ip);
  }, 10000);

  pc.ontrack = e => {
    if (PCS[ip] !== pc) return;
    if (e.streams && e.streams[0]) {
      STREAMS[ip] = e.streams[0];
      attachVideo(ip);
    }
  };

  pc.onconnectionstatechange = () => {
    // A superseded connection still fires 'closed' when we close it. Acting on
    // that event was what turned every restart into two parallel reconnect
    // chains, then four, until the browser ran out of memory.
    if (PCS[ip] !== pc) return;
    updateLiveBadge(ip);
    const state = pc.connectionState;
    if (state === 'failed' || state === 'disconnected' || state === 'closed') {
      delete PCS[ip];
      killPC(pc);
      const s = STATUS[ip];
      if (ST[ip] === 'live' && ((s && s.online) || isCamera(ip))) scheduleRetry(ip, streamName, 3000);
    }
  };

  // Keep-alive: send a dummy data channel ping every 15s to prevent timeout
  const dc = pc.createDataChannel('keepalive');
  pc._dc = dc;
  dc.onopen = () => {
    pc._keepAliveTimer = setInterval(() => {
      if (dc.readyState === 'open') dc.send('ping');
      else { clearInterval(pc._keepAliveTimer); pc._keepAliveTimer = null; }
    }, 15000);
  };
  dc.onclose = () => { if (pc._keepAliveTimer) { clearInterval(pc._keepAliveTimer); pc._keepAliveTimer = null; } };

  pc.addTransceiver('video', {direction: 'recvonly'});
  pc.addTransceiver('audio', {direction: 'recvonly'});

  try {
    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);

    // Wait for ICE gathering to complete (local candidates only = fast)
    await new Promise(resolve => {
      if (pc.iceGatheringState === 'complete') { resolve(); return; }
      const done = () => {
        if (pc._iceCheck) { pc.removeEventListener('icegatheringstatechange', pc._iceCheck); pc._iceCheck = null; }
        if (pc._iceTimeout) { clearTimeout(pc._iceTimeout); pc._iceTimeout = null; }
        resolve();
      };
      pc._iceCheck = () => { if (pc.iceGatheringState === 'complete') done(); };
      pc.addEventListener('icegatheringstatechange', pc._iceCheck);
      pc._iceTimeout = setTimeout(done, 2000); // max 2s wait
    });

    // Another start may have replaced us while we were awaiting.
    if (PCS[ip] !== pc) { killPC(pc); return; }

    const resp = await fetch('http://localhost:' + G2PORT + '/api/webrtc?src=' + encodeURIComponent(streamName), {
      method: 'POST',
      headers: {'Content-Type': 'application/sdp'},
      body: pc.localDescription.sdp
    });

    if (!resp.ok) throw new Error('SDP error ' + resp.status);
    const sdp = await resp.text();
    if (PCS[ip] !== pc) { killPC(pc); return; }
    await pc.setRemoteDescription({type: 'answer', sdp});
  } catch(e) {
    if (!String(e).includes('500')) console.warn('WebRTC ' + ip + ':', e.message||e);
    if (PCS[ip] === pc) delete PCS[ip];
    killPC(pc); // was missing — a failed handshake used to leak the whole pc
    if (ST[ip] === 'live') scheduleRetry(ip, streamName, 3000);
  }
}

function stopWebRTC(ip) {
  if (RETRY_TIMERS[ip]) { clearTimeout(RETRY_TIMERS[ip]); delete RETRY_TIMERS[ip]; }
  if (VERBINDE_TIMER[ip]) { clearTimeout(VERBINDE_TIMER[ip]); delete VERBINDE_TIMER[ip]; }
  const pc = PCS[ip];
  if (pc) {
    delete PCS[ip]; // drop the reference first so the close handler stays quiet
    killPC(pc);
  }
  delete STREAMS[ip];
  const video = document.getElementById('vid-' + ip);
  if (video) {
    video.srcObject = null;
    try { video.removeAttribute('src'); video.load(); } catch(e) {}
  }
  PAUSED[ip] = false;
  updateLiveBadge(ip);
}

function togglePauseWebRTC(ip) {
  const video = document.getElementById('vid-' + ip);
  if (!video || !video.srcObject) return;
  if (video.paused) {
    video.play();
    PAUSED[ip] = false;
  } else {
    video.pause();
    PAUSED[ip] = true;
  }
  updateLiveBadge(ip);
}

function updateLiveBadge(ip) {
  const badge = document.querySelector('.tile[data-ip="' + ip + '"] .tile-live-badge');
  if (!badge) return;
  const pc = PCS[ip];
  const isPaused = PAUSED[ip];
  const s = STATUS[ip];
  const isConnected = pc && (pc.connectionState === 'connected' || pc.connectionState === 'connecting');
  // Printer offline: MQTT connected but last_seen > 60s ago, or online===false
  const printerOffline = s && !s.online;

  badge.className = 'tile-live-badge'; // reset
  badge.classList.remove('dimmed');

  if (isPaused) {
    badge.textContent = 'PAUSE';
    badge.classList.add('dimmed');
    _setVideoActive(ip, true);
  } else if (printerOffline) {
    badge.textContent = 'OFFLINE';
    badge.classList.add('offline');
    _setVideoActive(ip, false);  // Video erlischen
  } else if (!G2online || !isConnected) {
    badge.textContent = 'LIVE';
    badge.classList.add('dim');
    _setVideoActive(ip, true);
  } else {
    badge.textContent = 'LIVE';
    _setVideoActive(ip, true);
  }
}

// Shows or hides the video (black when offline)
function _setVideoActive(ip, active) {
  const video = document.getElementById('vid-' + ip);
  if (!video) return;
  if (!active) {
    video.classList.add('fading');
    // Zeige Offline-Overlay
    const cam = video.closest('.tile-cam');
    if (cam && !cam.querySelector('.offline-overlay')) {
      const ov = document.createElement('div');
      ov.className = 'offline-overlay cam-idle';
      ov.innerHTML = '<div class="ci">📵</div><div class="cip">OFFLINE</div>';
      cam.appendChild(ov);
    }
  } else {
    video.classList.remove('fading');
    const cam = video.closest('.tile-cam');
    if (cam) {
      const ov = cam.querySelector('.offline-overlay');
      if (ov) ov.remove();
    }
  }
}

// ─── STATE ────────────────────────────────────────────────────────────────────
let P = [], ST = {}, SNAPS = {}, STATUS = {}, ERRORS = [], AF = new Set(), SQ = '', G2online = false, ONLINE_ONLY = false, OFFLINE_ONLY = false, FAVS_ONLY = false;
let FERTIG_ONLY = false;
// Remember the overview sorting (grid + list) across restarts.
let LIST_SORT = (() => { try { const s = JSON.parse(localStorage.getItem('listSort')); if (s && s.key) return {key:s.key, dir:(s.dir<0?-1:1)}; } catch(e){} return {key:'name', dir:1}; })();
let SYNC_SORT = {key:'name', dir:1};
let QL_SORT = {key:'name', dir:1};
const G2PORT = 1984;
const MODEL_PALETTE = ['#a78bfa','#00e5ff','#ff8c00','#39ff7e','#ff6b9d','#ffd700','#4fc3f7','#ff5252','#69ff47','#b39ddb'];
const _mcc = {};
function modelColor(m) {
  if (!_mcc[m]) { let h=0; for(let i=0;i<m.length;i++) h=(h*31+m.charCodeAt(i))>>>0; _mcc[m]=MODEL_PALETTE[h%MODEL_PALETTE.length]; }
  return _mcc[m];
}

// ─── INIT ─────────────────────────────────────────────────────────────────────
const BUILD_ID = '__BUILD_ID__';
const BUILD_MARKER = 'BUILD ' + BUILD_ID;
console.log('%cDruckerfarm ' + BUILD_MARKER + ' geladen', 'color:#0a0;font-weight:bold');
// Presence channel: while this connection is open, the server knows
// the window is there. On close it breaks -> the app shuts down immediately and
// cleanly (including go2rtc). On a reload the browser rebuilds it by
// itself. Not on the broadcast page (/broadcast) — that should not shut anything down.
function startAliveChannel(){
  if (window.BROADCAST_OPTS) return;
  try { new EventSource('/api/alive'); } catch(e) {}
}
document.addEventListener('DOMContentLoaded', async () => {
  if (window.BROADCAST_OPTS) { await bootBroadcast(); return; }
  startAliveChannel();
  purgeLegacyColors();
  // Settings first (language/filter), but an error must not
  // abbrechen.
  try { await loadSettings(); } catch(e) {}
  applyLang();
  loadMe(); // who is logged in → sets role, shows/hides admin controls

  // IMPORTANT: show the overview IMMEDIATELY — before printers/cameras are loaded.
  // Previously it loaded first and then showed; if a fetch hung
  // during the slow Windows startup phase, the UI was never
  // sichtbar (leere weisse Flaeche trotz laufendem Server).
  setPillFilter(START_FILTER || 'all');
  sv('uebersicht');
  markModeButton('btn-snap2');

  // Load data non-blocking; loadPrinters catches up by itself.
  loadPrinters();
  loadCameras().catch(()=>{});

  // What delays the start runs alongside. Previously starting the
  // images waited on the component check — which queries GitHub and can take
  // seconds. Until then the UI was frozen.
  chkG2();
  showG2Processes();
  tickers.push(setInterval(chkG2, 15000));
  tickers.push(setInterval(showG2Processes, 15000));
  loadErrors();
  tickers.push(setInterval(loadErrors, 5000));
  loadJobs();
  tickers.push(setInterval(loadJobs, 3000));
  loadPresence();
  tickers.push(setInterval(loadPresence, 30000));
  loadBoardTable();
  tickers.push(setInterval(loadStatus, 30000));

  // Fetch the status first, then draw: otherwise "Online" filters against empty
  // data and not a single tile appears — exactly why one
  // had to click the filter twice.
  // If the server is still paused, that also applies to the UI.
  try {
    const np = await (await fetch('/api/network/pause', {cache:'no-store'})).json();
    if (np && np.pausiert) showNetPause(true);
  } catch(e) {}

  await loadStatus();

  checkComponentsOnStart().then(() => {
    autoStartSnapshots();
    loadUpdateConfig();
    checkComponentUpdates(false);
    checkUpdate(false);
  });
});

// autoStartSnapshots starts the 2s view by itself once both tools
// are present. If ffmpeg is missing, only the hint remains — otherwise 42
// tiles with the same error would show.
function autoStartSnapshots() {
  const have = k => COMPONENTS.some(c => c.key === k && c.installed);
  if (!have('go2rtc') || !have('ffmpeg')) {
    SNAP_AUTO = false;
    return;
  }
  SNAP_MODE = 2;
  SNAP_AUTO = true;
  markModeButton('btn-snap2');
  renderGrid();
}

// ─── SERVER API ───────────────────────────────────────────────────────────────
// loadPrinters fetches the stored printer list. On start the server
// (Windows first scan) can be briefly slow — if the call fails or
// nothing comes, it retries several times instead of showing an empty grid forever.
async function loadPrinters(retry) {
  retry = retry || 0;
  try {
    const r = await fetch('/api/printers', {cache:'no-store'});
    if (!r.ok) throw new Error('HTTP '+r.status);
    const list = await r.json();
    if (Array.isArray(list)) {
      P = list;
      P.forEach(p=>{if(!ST[p.ip])ST[p.ip]='idle';});
      render();
      // Empty although just started? Server maybe not ready yet —
      // ein paar Mal erneut versuchen.
      if (P.length === 0 && retry < 10) setTimeout(()=>loadPrinters(retry+1), 1500);
    } else if (retry < 20) {
      setTimeout(()=>loadPrinters(retry+1), 1500);
    }
  } catch(e) {
    if (retry < 20) setTimeout(()=>loadPrinters(retry+1), 1500);
  }
}
async function apiAdd(p) { const r=await fetch('/api/printers',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(p)}); if(!r.ok)throw new Error(await r.text()); return r.json(); }
async function apiDelete(ip) { const r=await fetch('/api/printers/'+encodeURIComponent(ip),{method:'DELETE'}); if(!r.ok)throw new Error(await r.text()); }
async function apiPatch(ip,data) {
  const r = await fetch('/api/printers/'+encodeURIComponent(ip),{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)});
  if (!r.ok) throw new Error(await r.text());
}
async function apiImport(csv) { const r=await fetch('/api/import',{method:'POST',body:csv}); return r.json(); }

// ─── STATUS ───────────────────────────────────────────────────────────────────


async function loadStatus() {
  if (NETZ_PAUSE) return;
  // Self-healing: if the printer list is (still) empty, reload here. So
  // the grid fills at the latest on the next status tick, if the
  // first fetch at start failed.
  if (!P.length) loadPrinters();
  busyStart('fragt Drucker ab …');
  try {
    const r = await fetch('/api/status');
    const data = await r.json();
    STATUS = data || {};
    checkFarmAlarms();
    P.forEach(p => {
      refreshTile(p.ip);
      if (ST[p.ip] === 'live') updateLiveBadge(p.ip);
    });
    updStats();
    // Newly visible printers need a tile — not just a
    // toggled display. applyOnlineFilter only shows/hides existing
    // tiles; whoever was still offline at the last draw had
    // none. So it redraws as soon as the visible set
    // aendert.
    if (window.BROADCAST_OPTS) {
      const now2 = broadcastList().map(p => p.ip).join(',');
      if (now2 !== BC_LAST) renderBroadcast();
    }
    const jetzt = getFlt().map(p => p.ip).slice().sort().join(',');
    if (jetzt !== ZULETZT_SICHTBAR) {
      ZULETZT_SICHTBAR = jetzt;
      if (CURRENT_SUBVIEW === 'grid') renderGrid();
      else if (CURRENT_SUBVIEW === 'list') renderList();
    } else if (ONLINE_ONLY) {
      applyOnlineFilter();
    }
  } catch(e) {} finally { busyEnd(); }
}

async function loadErrors() {
  if (NETZ_PAUSE) return;
  try {
    const r = await fetch('/api/errors');
    ERRORS = await r.json() || [];
    // Update only status bars — never touch iframes
    P.forEach(p => refreshTile(p.ip));
  } catch(e) {}
}

// ─── JOBS / WORKER PANEL ──────────────────────────────────────────────────────
//
// A shared panel shows all running work: the server operations
// (search, delete, file sync) from /api/jobs plus the running upload from the
// browser. Each job has progress and can be cancelled. The header
// carries a counter of active jobs.
let JOBS = [];
let JOBS_OFFEN = false;

async function loadPresence() {
  try {
    const d = await (await fetch('/api/praesenz', {cache:'no-store'})).json();
    const b = document.getElementById('praesenz-banner');
    if (!b) return;
    if (d && d.andere && d.andere.length) {
      b.textContent = '⚠ ' + t('parallelPC') + ': ' + d.andere.join(', ');
      b.style.display = 'block';
    } else {
      b.style.display = 'none';
    }
  } catch(e) {}
}

async function loadJobs() {
  if (NETZ_PAUSE) return;
  try {
    const d = await (await fetch('/api/jobs', {cache:'no-store'})).json();
    JOBS = (d && d.jobs) || [];
  } catch(e) { JOBS = JOBS || []; }
  updateJobsBadge();
  if (JOBS_OFFEN) renderJobsPanel();
}

function jobLabel(kind, query) {
  if (kind === 'search') return t('jobSearch') + (query ? ': ' + query : '');
  if (kind === 'delete') return t('jobDelete');
  if (kind === 'sync')   return t('jobSync');
  if (kind === 'upload') return t('jobUpload');
  return kind;
}

// aktiveJobs counts what is actually working right now — server jobs plus upload.
function aktiveJobs() {
  let n = JOBS.filter(j => j.running).length;
  if (typeof uploadRunning === 'function' && uploadRunning()) n++;
  return n;
}

function updateJobsBadge() {
  const b = document.getElementById('jobs-badge');
  if (!b) return;
  const n = aktiveJobs();
  b.textContent = n;
  b.style.display = n > 0 ? 'inline-flex' : 'none';
}

function toggleJobsPanel() {
  const p = document.getElementById('jobs-panel');
  if (!p) return;
  JOBS_OFFEN = (p.style.display === 'none' || !p.style.display);
  p.style.display = JOBS_OFFEN ? 'flex' : 'none';
  if (JOBS_OFFEN) { loadJobs(); renderJobsPanel(); }
}

function jobRow(o) {
  const pct = o.total > 0 ? Math.min(100, Math.round(o.done / o.total * 100)) : (o.running ? 100 : 0);
  const fertig = !o.running;
  const fehler = (o.failed || 0) > 0;
  let sub = o.done + ' / ' + (o.total || '?');
  if (o.kind === 'delete' && o.deleted !== undefined) sub = o.deleted + ' ' + t('jobDeleted');
  if (fehler) sub += ' · ' + o.failed + ' ' + t('jobFailed');
  sub += ' · ' + (fertig ? t('jobDone') : (o.stopped ? t('jobPaused') : t('jobRunning')));
  const fillCls = fehler ? 'failed' : (fertig ? 'done' : '');
  const abbrechen = o.running
    ? '<button class="job-cancel" onclick="abbrechenJob(\'' + esc(String(o.id)) + '\')">✕</button>'
    : '';
  return '<div class="job-item' + (fertig ? ' done' : '') + '">'
    + '<div class="job-top"><span class="job-label">' + esc(jobLabel(o.kind, o.query)) + '</span>' + abbrechen + '</div>'
    + '<div class="job-bar"><div class="job-fill ' + fillCls + '" style="--w:' + pct + '%"></div></div>'
    + '<div class="job-sub">' + esc(sub) + '</div>'
    + '</div>';
}

function renderJobsPanel() {
  const el = document.getElementById('jobs-list');
  if (!el) return;
  let html = '';
  // Laufender Upload aus dem Browser zuerst.
  if (typeof uploadRunning === 'function' && uploadRunning() && typeof UPLOAD === 'object' && UPLOAD) {
    const pct = UPLOAD.total > 0 ? Math.round(UPLOAD.done / UPLOAD.total * 100) : 0;
    html += '<div class="job-item">'
      + '<div class="job-top"><span class="job-label">' + esc(jobLabel('upload')) + '</span>'
      + '<button class="job-cancel" onclick="cancelUpload()">✕</button></div>'
      + '<div class="job-bar"><div class="job-fill" style="--w:' + pct + '%"></div></div>'
      + '<div class="job-sub">' + UPLOAD.done + ' / ' + UPLOAD.total + ' · ' + esc(UPLOAD.note || t('jobRunning')) + '</div>'
      + '</div>';
  }
  html += JOBS.map(jobRow).join('');
  el.innerHTML = html || '<div class="jobs-empty">' + t('jobsEmpty') + '</div>';
}

async function abbrechenJob(id) {
  try {
    if (id === 'sync') {
      await fetch('/api/sync/stop', {method:'POST'});
    } else {
      await fetch('/api/job/stop?id=' + encodeURIComponent(id), {method:'POST'});
    }
  } catch(e) {}
  setTimeout(loadJobs, 300);
}

// ─── JOB-PLANER (Matrix Figur × Drucker-Modell) ───────────────────────────────
let PLAN_FIGS = [], PLAN_MODELS = [];
async function loadPlan() {
  try {
    const d = await (await fetch('/api/plan', {cache:'no-store'})).json();
    PLAN_FIGS = (d.figures||[]).map(f => ({id:f.id||'', name:f.name||'', file:f.file||'', qty:f.qty||{}}));
    PLAN_MODELS = d.models||[];
  } catch(e) { PLAN_FIGS = []; PLAN_MODELS = []; }
  renderPlan();
}
function renderPlan() {
  const el = byId('plan-matrix'); if(!el) return;
  if(!PLAN_MODELS.length){ el.innerHTML = '<div class="plan-empty">'+t('planNoModels')+'</div>'; return; }
  if(!PLAN_FIGS.length){ el.innerHTML = '<div class="plan-empty">'+t('planNoFig')+'</div>'; return; }
  let h = '<table class="plan-table"><thead><tr><th class="plan-fig-h">'+t('planFigure')+'</th>';
  PLAN_MODELS.forEach(m => { h += '<th>'+esc(m)+'</th>'; });
  h += '<th class="plan-sum-h">'+t('planSum')+'</th><th></th></tr></thead><tbody>';
  PLAN_FIGS.forEach((f,i) => {
    h += '<tr><td class="plan-fig"><input class="plan-name repo-input" value="'+esc(f.name)+'" oninput="planSetName('+i+',this.value)"></td>';
    let sum = 0;
    PLAN_MODELS.forEach(m => { const v = f.qty[m]||0; sum += v;
      h += '<td><input type="number" min="0" class="plan-qty" value="'+(v||'')+'" onchange="planSetQty('+i+',\''+esc(m)+'\',this.value)"></td>'; });
    h += '<td class="plan-sum">'+sum+'</td>';
    h += '<td><button class="pc-btn danger" onclick="planDelFigure('+i+')">'+t('delete')+'</button></td></tr>';
  });
  h += '</tbody><tfoot><tr><td class="plan-fig-h">'+t('planSum')+'</td>';
  let grand = 0;
  PLAN_MODELS.forEach(m => { let c = 0; PLAN_FIGS.forEach(f => c += (f.qty[m]||0)); grand += c; h += '<td class="plan-sum">'+c+'</td>'; });
  h += '<td class="plan-sum">'+grand+'</td><td></td></tr></tfoot></table>';
  el.innerHTML = h;
}
function planSetName(i,v){ if(PLAN_FIGS[i]) PLAN_FIGS[i].name = v; }
function planSetQty(i,m,v){ if(!PLAN_FIGS[i]) return; const n = parseInt(v)||0; if(n>0) PLAN_FIGS[i].qty[m] = n; else delete PLAN_FIGS[i].qty[m]; renderPlan(); }
function planAddFigure(){
  const nm = ((byId('plan-new-name')||{}).value||'').trim();
  if(!nm){ toast(t('planNameMissing'),'er'); return; }
  PLAN_FIGS.push({id:'', name:nm, file:'', qty:{}});
  byId('plan-new-name').value = '';
  renderPlan();
}
function planDelFigure(i){ PLAN_FIGS.splice(i,1); renderPlan(); }
async function savePlan(){
  const figures = PLAN_FIGS.filter(f => (f.name||'').trim()).map(f => ({id:f.id||'', name:f.name.trim(), file:f.file||'', qty:f.qty||{}}));
  try{
    const r = await fetch('/api/plan',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({figures})});
    if(!r.ok) throw new Error(await r.text());
    toast(t('planSaved'),'ok'); loadPlan();
  }catch(e){ toast(e.message||String(e),'er'); }
}

// ─── ANMELDUNG / BENUTZER / TOKENS ────────────────────────────────────────────
let MY_ROLE = '';
async function loadMe(){
  try{
    const d = await (await fetch('/api/me',{cache:'no-store'})).json();
    if(d.anon){ location.reload(); return; }
    MY_ROLE = d.role||'';
    const nm=byId('user-name'); if(nm) nm.textContent = (d.name||'').replace(/^token:/,'');
    const rl=byId('user-role'); if(rl) rl.textContent = roleLabel(MY_ROLE);
    document.body.classList.remove('role-viewer','role-operator','role-admin');
    if(MY_ROLE) document.body.classList.add('role-'+MY_ROLE);
    // Local access has no login — hide the whole user chip. Show it only for a
    // real remote session (name + role + logout).
    const chip=byId('user-chip'); if(chip) chip.style.display = d.local ? 'none' : '';
    if(MY_ROLE==='admin'){ loadUsers(); loadTokens(); }
  }catch(e){}
}
function roleLabel(r){ return r==='admin'?t('roleAdmin'):r==='operator'?t('roleOperator'):r==='viewer'?t('roleViewer'):r; }
async function doLogout(){
  try{ await fetch('/api/logout',{method:'POST'}); }catch(e){}
  location.href='/';
}
async function loadUsers(){
  const el=byId('users-list'); if(!el) return;
  try{
    const d=await (await fetch('/api/users',{cache:'no-store'})).json();
    const us=d.users||[];
    el.innerHTML = us.map(u=>{
      return '<div class="acc-row">'
        + '<span class="acc-name">'+esc(u.name)+'</span>'
        + '<select class="set-select acc-role" onchange="setUserRole(\''+esc(u.name)+'\',this.value)">'
        +   ['viewer','operator','admin'].map(r=>'<option value="'+r+'"'+(u.role===r?' selected':'')+'>'+roleLabel(r)+'</option>').join('')
        + '</select>'
        + '<button class="pc-btn" onclick="resetUserPw(\''+esc(u.name)+'\')">'+t('accessPw')+'</button>'
        + '<button class="pc-btn danger" onclick="delUser(\''+esc(u.name)+'\')">'+t('delete')+'</button>'
        + '</div>';
    }).join('') || '<div class="acc-empty">–</div>';
  }catch(e){}
}
async function addUser(){
  const name=(byId('new-user-name')||{}).value||'';
  const password=(byId('new-user-pass')||{}).value||'';
  const role=(byId('new-user-role')||{}).value||'viewer';
  if(!name.trim()||password.length<6){ toast(t('accessNeedPw'),'er'); return; }
  try{
    const r=await fetch('/api/users',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name:name.trim(),password,role})});
    if(!r.ok) throw new Error((await r.json().catch(()=>({}))).error||'Fehler');
    byId('new-user-name').value=''; byId('new-user-pass').value='';
    toast(t('accessSaved'),'ok'); loadUsers();
  }catch(e){ toast(e.message||String(e),'er'); }
}
async function setUserRole(name,role){
  try{
    const r=await fetch('/api/user?name='+encodeURIComponent(name),{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({role})});
    if(!r.ok) throw new Error((await r.json().catch(()=>({}))).error||'Fehler');
    toast(t('accessSaved'),'ok');
  }catch(e){ toast(e.message||String(e),'er'); loadUsers(); }
}
async function resetUserPw(name){
  const pw=prompt(t('accessNewPw')+' — '+name); if(pw===null) return;
  if(pw.length<6){ toast(t('accessNeedPw'),'er'); return; }
  try{
    const r=await fetch('/api/user?name='+encodeURIComponent(name),{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({password:pw})});
    if(!r.ok) throw new Error((await r.json().catch(()=>({}))).error||'Fehler');
    toast(t('accessSaved'),'ok');
  }catch(e){ toast(e.message||String(e),'er'); }
}
async function delUser(name){
  if(!confirm(t('accessDelUser')+' "'+name+'"?')) return;
  try{
    const r=await fetch('/api/user?name='+encodeURIComponent(name),{method:'DELETE'});
    if(!r.ok) throw new Error((await r.json().catch(()=>({}))).error||'Fehler');
    loadUsers();
  }catch(e){ toast(e.message||String(e),'er'); }
}
async function loadTokens(){
  const el=byId('tokens-list'); if(!el) return;
  try{
    const d=await (await fetch('/api/tokens',{cache:'no-store'})).json();
    const ts=d.tokens||[];
    el.innerHTML = ts.map(x=>{
      return '<div class="acc-row">'
        + '<span class="acc-name">'+esc(x.name)+'</span>'
        + '<span class="acc-role-tag">'+roleLabel(x.role)+'</span>'
        + '<button class="pc-btn danger" onclick="delToken(\''+esc(x.id)+'\')">'+t('delete')+'</button>'
        + '</div>';
    }).join('') || '<div class="acc-empty">–</div>';
  }catch(e){}
}
async function addToken(){
  const name=(byId('new-token-name')||{}).value||'';
  const role=(byId('new-token-role')||{}).value||'viewer';
  if(!name.trim()){ toast(t('accessNeedName'),'er'); return; }
  try{
    const r=await fetch('/api/tokens',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name:name.trim(),role})});
    if(!r.ok) throw new Error((await r.json().catch(()=>({}))).error||'Fehler');
    const d=await r.json();
    byId('new-token-name').value='';
    const rv=byId('token-reveal');
    if(rv){ rv.style.display='block'; rv.innerHTML='<b>'+t('accessTokenNew')+':</b> <code>'+esc(d.token)+'</code><br><span class="note-sm">'+t('accessTokenOnce')+'</span>'; }
    loadTokens();
  }catch(e){ toast(e.message||String(e),'er'); }
}
async function delToken(id){
  if(!confirm(t('accessDelToken')+'?')) return;
  try{
    const r=await fetch('/api/token?id='+encodeURIComponent(id),{method:'DELETE'});
    if(!r.ok) throw new Error('Fehler');
    loadTokens();
  }catch(e){ toast(e.message||String(e),'er'); }
}

// ─── AUFWACHEN NACH RUHEMODUS ─────────────────────────────────────────────────
//
// After standby or long minimizing, queries and video streams hang,
// although the server keeps running. We then do NOT reload and show NO
// overlay — in practice that only annoyed and, on brief dropouts,
// fired for no reason. Instead we quietly refresh the data in the
// background. If a query fails (server briefly gone), the
// existing error handling applies; once the server is back,
// the display fills itself on the next tick.
let LETZTER_TAKT = Date.now();
let GROSSE_LUECKE = false;

setInterval(() => {
  const now = Date.now();
  const luecke = now - LETZTER_TAKT;
  LETZTER_TAKT = now;
  if (luecke > 60000) {
    if (!document.hidden) sanfteErholung();
    else GROSSE_LUECKE = true; // beim Sichtbarwerden nachholen
  }
}, 5000);

document.addEventListener('visibilitychange', () => {
  if (document.hidden || NETZ_PAUSE) return;
  if (GROSSE_LUECKE) { GROSSE_LUECKE = false; sanfteErholung(); }
});
window.addEventListener('online', () => { if (!NETZ_PAUSE) sanfteErholung(); });

// sanfteErholung refetches the data and reattaches the videos — without
// a page switch, without an overlay.
function sanfteErholung() {
  if (NETZ_PAUSE) return;
  try {
    loadStatus();
    loadErrors();
    chkG2();
    showG2Processes();
    loadJobs();
    if (typeof resumeMedia === 'function') resumeMedia();
  } catch(e){}
}

// ─── go2rtc ───────────────────────────────────────────────────────────────────
async function chkG2() {
  if (NETZ_PAUSE) return;
  try {
    const r=await fetch('http://localhost:'+G2PORT+'/api/streams',{signal:AbortSignal.timeout(4000)});
    if(r.ok){
      const was=!G2online; G2online=true; G2_RESTARTING=false; G2_FAILS=0;
      if(G2_PREV!==true && CURRENT_SUBVIEW==='grid') { G2_PREV=true; setTimeout(()=>renderGrid(),50); } G2_PREV=true;
      const si=byId('setup-icon'); if(si) si.classList.remove('fehlt','arbeitet');
      if(was&&G2_RECOVERY_STREAMS.size>0){
        toast('go2rtc wiederhergestellt, starte '+G2_RECOVERY_STREAMS.size+' Streams…','ok');
        setTimeout(()=>{const back=Array.from(G2_RECOVERY_STREAMS).map(ip=>P.find(x=>x.ip===ip)).filter(Boolean);back.forEach(p=>setTileState(p.ip,activeMode()));G2_RECOVERY_STREAMS.clear();mountStreams(back);},1500);
      }
      setCls('sind','sind online');
      setText('stitle', t('g2running')+G2PORT);
      setText('sdesc', t('g2streams'));
      if(was){
        // go2rtc is back — continue in the currently selected mode.
        // Previously this was fixed to 'live': with the 2s default, favorites
        // trotzdem Videostreams aufgemacht.
        const resume=[];
        P.forEach(p=>{
          if(p.fav && (ST[p.ip]||'idle')==='idle') setTileState(p.ip,activeMode());
          if(ST[p.ip]==='live'||ST[p.ip]==='snap') resume.push(p);
          updateLiveBadge(p.ip);
        });
        if(resume.length) setTimeout(()=>mountStreams(resume),200);
      }
    } else throw 0;
  } catch {
    G2_FAILS++;
    if (G2_FAILS < 3) return;
    G2online=false;
    const si2=byId('setup-icon');
    if(si2){
      si2.classList.remove('fehlt','arbeitet');
      // Shortly after start or during a restart go2rtc is still working —
      // then pulsing yellow instead of red, so it is clear: it is starting up.
      const startphase = G2_RESTARTING || (Date.now()-APP_START < 45000);
      si2.classList.add(startphase ? 'arbeitet' : 'fehlt');
    }
    if(G2_PREV!==false && CURRENT_SUBVIEW==='grid') renderGrid(); // Wechsel online->offline
    G2_PREV=false;
    setCls('g2dot','g2dot offline');
    setCls('sind','sind offline');
    setText('stitle', t('g2offline'));
    setText('sdesc', t('g2noconn')+G2PORT);
    // Don't call renderGrid — just update badges
    P.forEach(p => { if(ST[p.ip]==='live') updateLiveBadge(p.ip); });
  }
}
async function restartG2() {
  // During the restart go2rtc is working — let the gear pulse yellow and
  // put the boxes into the waiting state instead of red/"offline" right away.
  G2_RESTARTING = true;
  const si = byId('setup-icon'); if (si) { si.classList.remove('fehlt'); si.classList.add('arbeitet'); }
  if (CURRENT_SUBVIEW === 'grid') renderGrid();
  try { await fetch('/api/go2rtc/restart', {method:'POST'}); } catch(e) {}
  toast(t('g2restarting'), 'ok');
  // The server needs a moment (stop, free the port, restart,
  // come up). So check repeatedly until go2rtc is back.
  let n = 0;
  const iv = setInterval(() => {
    n++;
    chkG2();
    showG2Processes();
    if (G2online || n >= 14) { clearInterval(iv); G2_RESTARTING = false; }
  }, 1500);
}

// showG2Processes fetches the current number of running go2rtc processes from the server.
async function showG2Processes() {
  try {
    const d = await (await fetch('/api/go2rtc/status',{cache:'no-store'})).json();
    const el = document.getElementById('g2-proc');
    if (el) {
      const n = d.prozesse || 0;
      const pids = Array.isArray(d.pids) ? d.pids : (d.pid ? [d.pid] : []);
      const pidTxt = pids.length ? ' · PID ' + pids.join(', ') : '';
      el.textContent = n + (n === 1 ? ' Prozess' : ' Prozesse') + pidTxt;
      el.classList.toggle('warn', n > 1);
    }
  } catch(e) {}
}

// killG2 is the emergency switch: stop all go2rtc processes, start one
// fresh. Asks first, because all videos briefly break.
async function killG2() {
  if (!confirm('Alle go2rtc-Prozesse beenden und einen frischen starten?\nVideos und Snapshots setzen dabei kurz aus.')) return;
  const btn = document.getElementById('btn-killg2');
  if (btn) { btn.disabled = true; btn.textContent = '⏳ beende …'; }
  try {
    const d = await (await fetch('/api/go2rtc/killall',{method:'POST'})).json();
    toast((d.beendet || 0) + ' go2rtc-Prozess(e) beendet, einer neu gestartet', 'ok');
  } catch(e) {
    toast('Killswitch fehlgeschlagen', 'er');
  } finally {
    if (btn) { btn.disabled = false; btn.textContent = '🧯 Alle go2rtc beenden'; }
    setTimeout(() => { chkG2(); showG2Processes(); }, 2500);
  }
}

// ─── STREAM CONTROLS ──────────────────────────────────────────────────────────

// activeMode returns the state a "running" tile must have in the currently
// selected display mode.
function activeMode() { return SNAP_MODE > 0 ? 'snap' : 'live'; }

// setTileState is the ONLY way to change a tile's state. Previously
// stopAll() and stopNonFavs() only set ST[ip]='idle' and re-rendered — the
// WebRTC connections kept running invisibly, because only the
// <video> elements were discarded. So a once-started
// stream could no longer be stopped.
function setTileState(ip, next) {
  const cur = ST[ip] || 'idle';
  if (cur === next) return;
  if (cur === 'live') stopWebRTC(ip);
  if (cur === 'snap') stopSnapPoll(ip);
  ST[ip] = next;
}

function setTileStates(printers, next) { printers.forEach(p => setTileState(p.ip, next)); }

function startAll() {
  if(!G2online){toast(t('toastG2offline'),'er');return;}
  const vis=getFlt();
  SNAP_AUTO = true;
  setTileStates(vis, activeMode());
  renderGrid();
  toast(vis.length+t('toastStarted'),'ok');
}
function stopAll() {
  SNAP_AUTO = false; // sonst zieht renderGrid die Kacheln sofort wieder hoch
  setTileStates(P, 'idle');
  renderGrid();
  toast(t('toastStopped'),'ok');
}
function startFavs() {
  if(!G2online){toast(t('toastG2offline'),'er');return;}
  const f=P.filter(p=>p.fav);
  if(!f.length){toast(t('toastNoFavs'),'er');return;}
  SNAP_AUTO = true;
  setTileStates(f, activeMode());
  renderGrid();
  toast(f.length+t('toastFavStarted'),'ok');
}
function stopNonFavs() {
  setTileStates(P.filter(p=>!p.fav), 'idle');
  renderGrid();
  toast(t('toastNonFavStopped'),'ok');
}
function tilePlay(ip) {
  if(!G2online && SNAP_MODE<=0){toast(t('toastG2offline'),'er');return;}
  setTileState(ip, activeMode());
  renderTile(ip);
  const p=P.find(x=>x.ip===ip);
  if(p) mountStreams([p]);
}
function tileStop(ip) { setTileState(ip,'idle'); renderTile(ip); }
// Show first, then save. Previously the UI waited on the
// server — which on save writes the entire configuration including
// upload history to disk. With a full history that was a noticeable
// delay on a click that does nothing but toggle a star.
// faerben.
function tileToggleFav(ip) {
  const p = P.find(x => x.ip === ip);
  if (!p) return;
  p.fav = !p.fav;

  // Make visible immediately — without rebuilding the tile, which would tear the
  // laufenden Stream ab.
  refreshTile(ip);
  updStats();
  toast(p.fav ? '★ ' + p.name + t('toastFavAdded') : p.name + t('toastFavRemoved'), 'ok');

  // Save in the background. If it fails, it is rolled back.
  apiPatch(ip, {fav: p.fav}).catch(() => {
    p.fav = !p.fav;
    refreshTile(ip);
    updStats();
    toast('Konnte ' + p.name + ' nicht speichern', 'er');
  });
}

// ─── RENDER ───────────────────────────────────────────────────────────────────
function render(){
  renderGrid();
  renderList();
  renderQL();
  updStats();
  // Update the FileSync printer list too — otherwise it shows the old state after
  // adding/removing a printer.
  if(typeof renderSyncPrinterList==='function') renderSyncPrinterList();
  if(typeof renderYaml==='function') renderYaml();
  if(typeof renderCamAdmin==='function') renderCamAdmin();
}

// Shows/hides tiles without a DOM rebuild (protects running streams)
function applyOnlineFilter() {
  document.querySelectorAll('#printerGrid .tile').forEach(tile => {
    if (tile.classList.contains('cam-tile')) { setShown(tile, true); return; } // Kameras immer zeigen
    const ip = tile.getAttribute('data-ip');
    const s = STATUS[ip];
    const isOnline = s && s.online;
    setShown(tile, isOnline);
  });
}

// isPaused: paused or cancelled. Both need the same attention —
// the machine is stopped and nobody notices.
function isPaused(ip){
  const s=STATUS[ip]; if(!s||!s.online) return false;
  const st=String(s.gcode_state||'').toUpperCase();
  return st==='PAUSE'||st==='PAUSED'||st==='FAILED'||(s.print_error>0);
}
// istFertig: a print reached 100 %. The FINISH state stays on the device
// until a new job starts — then it is RUNNING and the
// marker drops. The progress serves as a fallback if a model
// briefly at 100 % has not yet switched to FINISH.
function istFertig(ip){
  const s=STATUS[ip]; if(!s||!s.online) return false;
  const st=String(s.gcode_state||'').toUpperCase();
  if(st==='RUNNING'||st==='PAUSE'||st==='PAUSED') return false;
  return st==='FINISH'||(Number(s.progress)>=100);
}
function getFlt(){const f=P.filter(p=>{const ms=!SQ||p.name.toLowerCase().includes(SQ)||p.ip.includes(SQ);const mf=AF.size===0||AF.has(p.model);const online=!!(STATUS[p.ip]&&STATUS[p.ip].online);if(ONLINE_ONLY&&!online)return false;if(OFFLINE_ONLY&&online)return false;if(PAUSED_ONLY&&!isPaused(p.ip))return false;if(FAVS_ONLY&&!p.fav)return false;if(FERTIG_ONLY&&!istFertig(p.ip))return false;return ms&&mf;});return (typeof sortList==='function')?sortList(f):f;}
function sname(p){return(p.name||'').toLowerCase().replace(/[^a-z0-9]/g,'-').replace(/-+/g,'-').replace(/^-|-$/g,'')||'printer-'+p.ip.replace(/\./g,'-');}
function esc(s){return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');}


function buildDropdown(p) {
  const s = STATUS[p.ip];
  const errs = ERRORS ? ERRORS.find(e => e.ip === p.ip) : null;
  let rows = '';

  if (s && s.online) {
    // Print job info
    if (s.subtask_name) rows += '<div class="tt-row"><span class="tt-label">📄 Datei</span><span class="tt-val">' + esc(s.subtask_name.replace(/\.gcode$/i,'')) + '</span></div>';
    if (s.progress > 0) rows += '<div class="tt-row"><span class="tt-label">📊 Fortschritt</span><span class="tt-val">' + s.progress + '%</span></div>';
    if (s.remain_time > 0) rows += '<div class="tt-row"><span class="tt-label">⏱ Verbleibend</span><span class="tt-val">' + Math.round(s.remain_time) + ' min</span></div>';
    if (s.layer > 0 && s.total_layers > 0) rows += '<div class="tt-row"><span class="tt-label">📐 Schicht</span><span class="tt-val">' + s.layer + ' / ' + s.total_layers + '</span></div>';
    // Temperatures
    if (s.nozzle_temp > 0 || s.bed_temp > 0 || s.chamber_temp > 0) {
      if (rows) rows += '<div class="tt-sep"></div>';
      if (s.nozzle_temp > 0) rows += '<div class="tt-row"><span class="tt-label">🌡 Nozzle</span><span class="tt-val">' + Math.round(s.nozzle_temp) + '°' + (s.nozzle_target > 0 ? ' / ' + Math.round(s.nozzle_target) + '°' : '') + '</span></div>';
      if (s.bed_temp > 0)    rows += '<div class="tt-row"><span class="tt-label">🌡 Bett</span><span class="tt-val">'    + Math.round(s.bed_temp)    + '°' + (s.bed_target    > 0 ? ' / ' + Math.round(s.bed_target)    + '°' : '') + '</span></div>';
      if (s.chamber_temp > 0) rows += '<div class="tt-row"><span class="tt-label">🌡 Kammer</span><span class="tt-val">' + Math.round(s.chamber_temp) + '°</span></div>';
    }
  } else if (!s || !p.serial) {
    rows = '<div class="tt-empty">Keine Daten — Seriennummer eingetragen?</div>';
  } else {
    rows = '<div class="tt-empty">Offline</div>';
  }

  // Errors always shown if present
  if (errs && errs.messages && errs.messages.length) {
    if (rows) rows += '<div class="tt-sep"></div>';
    errs.messages.forEach(m => { rows += '<div class="tt-err">' + esc(m) + '</div>'; });
  }

  if (!rows) return '<div class="tile-dropdown" id="dd-'+esc(p.ip)+'"><div class="tt-empty">–</div></div>';
  return '<div class="tile-dropdown" id="dd-'+esc(p.ip)+'">' + rows + '</div>';
}

// buildCameraBtn: Kamera dieses Druckers dauerhaft an/aus (Privat).
function buildCameraBtn(p) {
  const aus = isCameraOff(p.ip);
  return '<div class="dd-sep"></div>'
    + '<button class="hbtn btn-sm ' + (aus ? 'green' : '') + '" onclick="toggleCamera(\'' + esc(p.ip) + '\')">'
    + (aus ? t('camTurnOn') : t('camTurnPrivate')) + '</button>';
}

// buildBlinkBtn: error blinking on/off for THIS printer — only on models
// that blink at all (no signal light).
function buildBlinkBtn(p) {
  if (!isBlinkModel(p.model)) return '';
  const aus = isBlinkOff(p.ip);
  return '<button class="hbtn btn-sm ' + (aus ? 'green' : '') + '" style="margin-top:6px" onclick="toggleBlink(\'' + esc(p.ip) + '\')">'
    + (aus ? '🔔 ' + t('blinkTurnOn') : '🔕 ' + t('blinkTurnOff')) + '</button>';
}

// ddRefresh rebuilds the content of the open detail box (after a toggle).
function ddRefresh(ip) {
  if (DD_OPEN_IP !== ip) return;
  const pop = document.getElementById('dd-pop');
  const p = P.find(x => x.ip === ip);
  if (pop && p) pop.innerHTML = '<div class="dd-title">' + esc(p.name || ip) + '<span class="dd-close" title="' + esc(t('close')) + '" onclick="event.stopPropagation();closeDD()">✕</span></div>' + buildDropdownContent(p);
}

async function toggleCamera(ip) {
  const aus = !isCameraOff(ip);
  try {
    const r = await fetch('/api/camera-off', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ip, aus})});
    if (!r.ok) throw new Error('HTTP ' + r.status);
    if (aus) { KAMERA_AUS[ip] = true; if (ST[ip]==='live') stopWebRTC(ip); if (ST[ip]==='snap') stopSnapPoll(ip); }
    else { delete KAMERA_AUS[ip]; }
    renderTile(ip); ddRefresh(ip); updateFSVCell(ip, printerFields(P.find(x=>x.ip===ip)||{ip}));
    if (!aus && (ST[ip]==='live' || ST[ip]==='snap')) { const p=P.find(x=>x.ip===ip); if (p) mountStreams([p]); }
    toast(aus ? t('camPrivMsg') : t('camOnMsg'), 'ok');
  } catch(e) { toast('Fehlgeschlagen', 'er'); }
}

async function toggleBlink(ip) {
  const aus = !isBlinkOff(ip);
  try {
    const r = await fetch('/api/blink', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ip, aus})});
    if (!r.ok) throw new Error('HTTP ' + r.status);
    if (aus) BLINK_AUS[ip] = true; else delete BLINK_AUS[ip];
    ddRefresh(ip);
    toast(aus ? t('blinkOffMsg') : t('blinkOnMsg'), 'ok');
  } catch(e) { toast('Fehlgeschlagen', 'er'); }
}

// buildFwUpdateBtn shows a button in the detail menu when a firmware update is
// pending for THIS printer — so it can be updated right there.
function buildFwUpdateBtn(p) {
  const ups = (STATUS[p.ip] && STATUS[p.ip].fw_updates) || [];
  if (!ups.length) return '';
  const geplant = !!(STATUS[p.ip] && STATUS[p.ip].fw_scheduled);
  return '<div class="dd-sep"></div>'
    + '<button class="hbtn btn-sm primary" onclick="startPrinterUpdate(\'' + esc(p.ip) + '\')">🔄 ' + t('fwInstall') + ' (' + ups.length + ')</button>'
    + '<button class="hbtn btn-sm ' + (geplant ? 'green' : '') + '" style="margin-top:6px" onclick="schedulePrinterUpdate(\'' + esc(p.ip) + '\',' + (!geplant) + ')">'
    + (geplant ? t('fwJobOn') : t('fwJobSet')) + '</button>';
}

// schedulePrinterUpdate: queue a firmware update as a job (starts automatically after
// the print) or cancel the queue. If the printer is idle, the
// update starts immediately — the server reports this (started:true).
async function schedulePrinterUpdate(ip, an) {
  try {
    const r = await fetch('/api/fwupdate/schedule', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ip, auto: an})});
    if (!r.ok) throw new Error('HTTP ' + r.status);
    const d = await r.json();
    if (STATUS[ip]) STATUS[ip].fw_scheduled = !!d.scheduled;
    ddRefresh(ip);
    toast(d.started ? t('fwStartMsg') : (d.scheduled ? t('fwJobMsg') : t('fwJobOffMsg')), 'ok');
  } catch(e) { toast('Fehlgeschlagen', 'er'); }
}

function buildDropdownContent(p) {
  const s = STATUS[p.ip];
  const errs = ERRORS ? ERRORS.find(e => e.ip === p.ip) : null;
  let rows = '';
  if (s && s.online) {
    { const stt = String(s.gcode_state||'').toUpperCase();
      const ack = isErrAcked(p.ip);
      const shownState = (ack && stt==='FAILED') ? '' : stt;
      const stLabel = {RUNNING:'PRINTING',PAUSE:'PAUSED',PAUSED:'PAUSED',FAILED:'FAILED',FINISH:'DONE',IDLE:'IDLE',PREPARE:'PREPARE',SLICING:'SLICING'}[shownState] || shownState;
      if (stLabel) rows += '<div class="dd-row"><span class="dd-lbl">'+t('mdStatus')+'</span><span class="dd-val '+(shownState==='FAILED'?'dd-val-failed':'')+'">'+esc(stLabel)+'</span></div>';
      // Acknowledge button directly after the status line.
      if (rawHatFehler(p.ip) && !ack) {
        rows += '<div class="dd-row"><button class="pc-btn" style="width:100%" onclick="event.stopPropagation();ackError(\''+esc(p.ip)+'\')">'+t('mdAckErr')+'</button></div>';
      } }
    // Colored print preview (plate thumbnail from the .3mf on the SD). Hidden
    // automatically when there is none for the current file.
    { const pf = printFileName(s);
      if (pf && /\.3mf$/i.test(pf)) {
        rows += '<div class="dd-preview" title="'+esc(t('mdPreview'))+'"><img class="dd-preview-img" loading="lazy" src="/api/preview?ip='+encodeURIComponent(p.ip)+'&file='+encodeURIComponent(pf)+'" onerror="this.closest(\'.dd-preview\').style.display=\'none\'" alt="'+esc(t('mdPreview'))+'"></div>';
      } }
    { const pfn = printFileName(s); if (pfn) rows += '<div class="dd-row"><span class="dd-lbl">Datei</span><span class="dd-val">'+esc(pfn)+'</span></div>'; }
    if (s.progress > 0)    rows += '<div class="dd-row"><span class="dd-lbl">Fortschritt</span><span class="dd-val">'+s.progress+'%</span></div>';
    if (s.remain_time > 0) {
      const rt=Math.round(s.remain_time);
      const dd=Math.floor(rt/1440),hh=Math.floor((rt%1440)/60),mm=rt%60;
      const rtStr=(dd>0?dd+'d ':'')+((hh>0||dd>0)?hh+'h ':'')+mm+'m';
      rows += '<div class="dd-row"><span class="dd-lbl">Verbleibend</span><span class="dd-val">'+rtStr+'</span></div>';
    }
    if (s.layer > 0 && s.total_layers > 0) rows += '<div class="dd-row"><span class="dd-lbl">Schicht</span><span class="dd-val">'+s.layer+' / '+s.total_layers+'</span></div>';
    if (s.nozzle_temp > 0 || s.bed_temp > 0 || s.chamber_temp > 0) {
      if (rows) rows += '<div class="dd-sep"></div>';
      if (s.nozzle_temp > 0)  rows += '<div class="dd-row"><span class="dd-lbl">Nozzle</span><span class="dd-val">'+Math.round(s.nozzle_temp)+'°'+(s.nozzle_target>0?' / '+Math.round(s.nozzle_target)+'°':'')+'</span></div>';
      if (s.bed_temp > 0)     rows += '<div class="dd-row"><span class="dd-lbl">Bett</span><span class="dd-val">'+Math.round(s.bed_temp)+'°'+(s.bed_target>0?' / '+Math.round(s.bed_target)+'°':'')+'</span></div>';
      if (s.chamber_temp > 0) rows += '<div class="dd-row"><span class="dd-lbl">Kammer</span><span class="dd-val">'+Math.round(s.chamber_temp)+'°</span></div>';
    }
    if (rows) rows += '<div class="dd-sep"></div>';
  rows += '<div class="dd-row"><span class="dd-lbl">IP</span><span class="dd-val mono">'+esc(p.ip)+'</span></div>';
  rows += deviceRows(p);
  if (!rows) rows = '<div class="dd-none">Idle — keine Daten</div>';
  } else {
    rows = '<div class="dd-none">'+(p.serial ? 'Offline' : 'Keine Seriennummer')+'</div>';
  }
  if (errs && errs.messages && errs.messages.length) {
    rows += '<div class="dd-sep"></div>';
    errs.messages.forEach(m => { rows += '<div class="dd-err">'+esc(m)+'</div>'; });
  }
  // Two columns so the menu does not grow endlessly downward: left the
  // details and the AMS, right the controls including temperature and filament.
  const strip = h => h.replace(/^<div class="dd-sep"><\/div>/, '');
  const flag = strip(buildRepairBtn(p) + buildCameraBtn(p) + buildBlinkBtn(p) + buildFwUpdateBtn(p));
  const ki = strip(buildXcamControls(p) + buildAIButtons(p));
  return '<div class="dd-blocks">'
    + '<div class="dd-block"><div class="dd-block-title">Status</div>' + rows + '</div>'
    + '<div class="dd-block"><div class="dd-block-title">Steuerung</div>' + buildPrintControls(p) + buildTempControls(p) + '</div>'
    + '<div class="dd-block"><div class="dd-block-title">'+t('ddFlag')+'</div>' + flag + '</div>'
    + '<div class="dd-block"><div class="dd-block-title">'+t('ddAI')+'</div>' + ki + '</div>'
    + '<div class="dd-block"><div class="dd-block-title">'+t('ddMedia')+'</div>' + buildMediaControls(p) + '</div>'
    + '<div class="dd-block"><div class="dd-block-title">AMS</div>' + buildAMSUnits(p) + '</div>'
    + '</div>';
}

// buildMediaControls: timelapse on/off (with a "keep off" option) and
// quick access to media management (videos/snapshots).
function buildMediaControls(p) {
  const s = STATUS[p.ip] || {};
  const online = !!s.online;
  let h = '';
  // Timelapse switch — state follows what the printer reports.
  const tlOn = s.timelapse_known ? (s.timelapse === 'enable') : false;
  const known = !!s.timelapse_known;
  h += '<div class="dd-row"><span class="dd-lbl">'+t('mediaTimelapse')+'</span><span class="dd-val">'
     + '<label class="tl-switch"><input type="checkbox" '+(tlOn?'checked':'')+' '+(online?'':'disabled')
     + ' onchange="toggleTimelapse(\''+esc(p.ip)+'\', this.checked)"> '
     + (known ? (tlOn ? t('on') : t('off')) : '?') + '</label></span></div>';
  // "Keep off" — the tool then regularly re-sends "off" if the
  // printer turns the timelapse back on by itself.
  const enf = !!(MEDIA_ENFORCE && MEDIA_ENFORCE[p.ip]);
  h += '<div class="dd-row"><span class="dd-lbl">'+t('mediaKeepOff')+'</span><span class="dd-val">'
     + '<label class="tl-switch"><input type="checkbox" '+(enf?'checked':'')+' '+(online?'':'disabled')
     + ' onchange="enforceTimelapse(\''+esc(p.ip)+'\', this.checked)"></label></span></div>';
  h += '<div class="dd-sep"></div>';
  h += '<div class="pc-row">'
     + '<button class="pc-btn" '+(online?'':'disabled')+' onclick="openMediaMgr(\''+esc(p.ip)+'\',\'video\')">'+t('mediaVideos')+'</button>'
     + '<button class="pc-btn" '+(online?'':'disabled')+' onclick="openMediaMgr(\''+esc(p.ip)+'\',\'snapshot\')">'+t('mediaSnapshots')+'</button>'
     + '</div>';
  if (!online) h += '<div class="dd-none mt5">'+t('offline')+'</div>';
  return h;
}

// toggleTimelapse turns the timelapse on/off. When turning off, an
// evtl. gesetzte „dauerhaft aus"-Markierung erhalten.
async function toggleTimelapse(ip, on) {
  const enforce = !!(MEDIA_ENFORCE && MEDIA_ENFORCE[ip]) && !on;
  try {
    const r = await fetch('/api/media/timelapse?ip='+encodeURIComponent(ip)+'&on='+(on?'1':'0')+'&enforce='+(enforce?'1':'0'), {method:'POST'});
    const d = await r.json();
    if (d.error) { toast(d.error, 'er'); return; }
    toast(t('mediaTimelapse')+': '+(on?t('on'):t('off')), 'ok');
  } catch(e) { toast(e.message, 'er'); }
}

// enforceTimelapse sets/removes the "keep off" marker. Enabled
// means: the tool regularly sends "off" if the printer turns the timelapse
// back on by itself.
async function enforceTimelapse(ip, keepOff) {
  if (!MEDIA_ENFORCE) MEDIA_ENFORCE = {};
  if (keepOff) MEDIA_ENFORCE[ip] = true; else delete MEDIA_ENFORCE[ip];
  try {
    // keepOff => timelapse off + enforce; otherwise only clear enforce (leave state).
    const on = keepOff ? '0' : (STATUS[ip] && STATUS[ip].timelapse === 'enable' ? '1' : '0');
    const r = await fetch('/api/media/timelapse?ip='+encodeURIComponent(ip)+'&on='+on+'&enforce='+(keepOff?'1':'0'), {method:'POST'});
    const d = await r.json();
    if (d.error) { toast(d.error, 'er'); return; }
    toast(t('mediaKeepOff')+': '+(keepOff?t('on'):t('off')), 'ok');
  } catch(e) { toast(e.message, 'er'); }
}

// ftpErrText returns the translated error message: prefers the stable
// key (error_code) via t(), otherwise the server's plain text. So
// e.g. the port-990 message appears in the currently set language.
function ftpErrText(d) {
  if (d && d.error_code) { const s = t(d.error_code); if (s && s !== d.error_code) return s; }
  return (d && d.error) ? d.error : '';
}

// ─── MEDIEN-VERWALTUNG (Overlay) ──────────────────────────────────────────────
let MEDIA_MGR = { ip:'', kind:'video' };

function ensureMediaOverlay() {
  let ov = document.getElementById('media-ov');
  if (ov) return ov;
  ov = document.createElement('div');
  ov.id = 'media-ov';
  ov.className = 'media-ov hidden';
  ov.innerHTML = '<div class="media-box">'
    + '<div class="media-head"><span id="media-title"></span>'
    + '<button class="media-x" onclick="closeMediaMgr()">✕</button></div>'
    + '<div id="media-info" class="media-info"></div>'
    + '<div id="media-list" class="media-list"></div>'
    + '<div class="media-foot">'
    + '<button class="pc-btn danger" id="media-clear" onclick="mediaClear()"></button>'
    + '<button class="pc-btn" onclick="loadMediaList()">'+t('mediaReload')+'</button>'
    + '</div></div>';
  ov.addEventListener('click', e => { if (e.target === ov) closeMediaMgr(); });
  document.body.appendChild(ov);
  return ov;
}

function openMediaMgr(ip, kind) {
  MEDIA_MGR = { ip: ip, kind: kind };
  const ov = ensureMediaOverlay();
  setShown(ov, true);
  loadMediaList();
}
function closeMediaMgr() { const ov = document.getElementById('media-ov'); if (ov) setShown(ov, false); }

async function loadMediaList() {
  const ip = MEDIA_MGR.ip, kind = MEDIA_MGR.kind;
  const p = (P.find(x=>x.ip===ip) || {name: ip});
  const title = document.getElementById('media-title');
  const info = document.getElementById('media-info');
  const list = document.getElementById('media-list');
  const clr = document.getElementById('media-clear');
  if (title) title.textContent = (kind==='video'?t('mediaVideos'):t('mediaSnapshots')) + ' — ' + (p.name||ip);
  if (clr) clr.textContent = (kind==='video'?t('mediaClearVideo'):t('mediaClearSnap'));
  if (list) list.innerHTML = '<div class="dd-none">'+t('mediaLoading')+'</div>';
  if (info) info.textContent = '';
  try {
    const d = await (await fetch('/api/media/list?ip='+encodeURIComponent(ip)+'&kind='+kind, {cache:'no-store'})).json();
    if (d.error) { list.innerHTML = '<div class="dd-err">'+esc(ftpErrText(d))+'</div>'; return; }
    const files = d.files || [];
    if (info && (d.free || d.total)) {
      let s = '';
      if (d.total) s += fmtSize(d.total)+' '+t('mediaTotal');
      if (d.free)  s += (s?' · ':'') + fmtSize(d.free)+' '+t('mediaFree');
      info.textContent = files.length + ' ' + t('mediaFiles') + (s ? '   ·   '+s : '');
    } else if (info) {
      info.textContent = files.length + ' ' + t('mediaFiles');
    }
    if (!files.length) { list.innerHTML = '<div class="dd-none">'+t('mediaEmpty')+'</div>'; return; }
    files.sort((a,b) => natCmp(a.name, b.name));
    let html = '';
    files.forEach(f => {
      const dl = '/api/media/download?ip='+encodeURIComponent(ip)+'&kind='+kind+'&path='+encodeURIComponent(f.path);
      html += '<div class="media-row">'
        + '<span class="media-name" title="'+esc(f.path)+'">'+esc(f.name)+'</span>'
        + '<span class="media-size mono">'+fmtSize(f.size)+'</span>'
        + '<a class="pc-btn sm" href="'+dl+'" target="_blank" rel="noopener">'+t('mediaDownload')+'</a>'
        + '<button class="pc-btn sm danger" onclick="mediaDelete(\''+esc(f.path)+'\')">'+t('mediaDelete')+'</button>'
        + '</div>';
    });
    list.innerHTML = html;
  } catch(e) {
    list.innerHTML = '<div class="dd-err">'+esc(e.message)+'</div>';
  }
}

async function mediaDelete(fullPath) {
  if (!confirm(t('mediaDelConfirm'))) return;
  const ip = MEDIA_MGR.ip, kind = MEDIA_MGR.kind;
  try {
    const r = await fetch('/api/media/delete?ip='+encodeURIComponent(ip)+'&kind='+kind+'&path='+encodeURIComponent(fullPath), {method:'DELETE'});
    const d = await r.json();
    if (d.error) { toast(d.error, 'er'); return; }
    toast(t('mediaDeleted'), 'ok');
    loadMediaList();
  } catch(e) { toast(e.message, 'er'); }
}

async function mediaClear() {
  const ip = MEDIA_MGR.ip, kind = MEDIA_MGR.kind;
  if (!confirm(t('mediaClearConfirm'))) return;
  const btn = document.getElementById('media-clear');
  if (btn) { btn.disabled = true; }
  try {
    const r = await fetch('/api/media/clear?ip='+encodeURIComponent(ip)+'&kind='+kind, {method:'POST'});
    const d = await r.json();
    if (d.error) { toast(ftpErrText(d), 'er'); }
    else toast(t('mediaCleared')+': '+(d.deleted||0), 'ok');
    loadMediaList();
  } catch(e) { toast(e.message, 'er'); }
  finally { if (btn) btn.disabled = false; }
}

// runFtpCheck prüft alle Drucker auf FTP-Erreichbarkeit (Port 990), Login und
// vorhandenen Speicher und zeigt je Drucker die Ursache. So lässt sich die Farm
// gezielt bereinigen, statt Drucker für Drucker durchzuprobieren.
async function runFtpCheck() {
  const el = document.getElementById('ftpcheck-result');
  const sum = document.getElementById('ftpcheck-sum');
  if (el) el.innerHTML = '<div class="dd-none">'+t('mediaLoading')+'</div>';
  if (sum) sum.textContent = '';
  try {
    const d = await (await fetch('/api/ftpcheck', {cache:'no-store'})).json();
    const res = d.results || [];
    // Einordnung je Zeile: grün = alles ok, gelb = verbunden/leer, rot = Problem.
    const cls = r => (r.port==='open' && r.login==='ok' && r.storage) ? 'ok'
      : (r.port==='open' && r.login==='ok' && !r.storage) ? 'warn' : 'err';
    const portTxt = { open:t('ftpPortOpen'), refused:t('ftpPortRefused'), timeout:t('ftpPortTimeout'), unreachable:t('ftpPortUnreach') };
    const rows = res.slice().sort((a,b)=>natCmp(a.name||a.ip,b.name||b.ip)).map(r => {
      let reason = r.code ? (t(r.code)!==r.code ? t(r.code) : r.reason) : r.reason;
      if (r.switched) reason = '↪ ' + t('ftpIpSwitched') + ' ' + r.ip + (reason ? ' · ' + reason : '');
      const badge = cls(r)==='ok' ? '<span class="ok">✓</span>' : (cls(r)==='warn' ? '<span class="warn">▲</span>' : '<span class="err">✗</span>');
      return '<div class="ftpc-row '+cls(r)+'">'
        + '<span class="ftpc-badge">'+badge+'</span>'
        + '<span class="ftpc-name">'+esc(r.name||r.ip)+'</span>'
        + '<span class="ftpc-port">'+(portTxt[r.port]||r.port)+'</span>'
        + '<span class="ftpc-reason">'+esc(reason||'')+'</span>'
        + '</div>';
    }).join('');
    const okN = res.filter(r=>cls(r)==='ok').length;
    const problems = res.length - okN;
    if (sum) sum.textContent = okN + ' ' + t('ftpCheckOk') + ' · ' + problems + ' ' + t('ftpCheckProblem')
      + ((d.ip_fixed>0) ? ' · ' + d.ip_fixed + ' ' + t('ftpIpFixed') : '');
    el.innerHTML = rows || '<div class="dd-none">—</div>';
  } catch(e) { if (el) el.innerHTML = '<div class="dd-err">'+esc(e.message)+'</div>'; }
}

// ─── FARMWEITE MEDIEN-AKTIONEN (File Sync) ────────────────────────────────────
// The target set is the printers checked in the File Sync list; if none
// is checked, the actions apply to ALL printers.
function farmMediaTargets() {
  const sel = syncSelectedPrinters();
  return (sel.length ? sel : P).map(p => p.ip);
}

function farmMediaResult(results) {
  const el = document.getElementById('sync-media-result');
  if (!el) return;
  const rows = (results||[]).map(r => {
    const ok = r.ok && !r.error;
    const info = r.error ? esc(r.error) : ((r.deleted!==undefined? r.deleted+' '+t('mediaCleared') : t('mediaDone')||'ok'));
    return '<div class="media-row"><span class="media-name">'+esc(r.name||r.ip)+'</span>'
      + '<span class="'+(ok?'':'sd-err')+'">'+(ok?'✓ ':'✗ ')+info+'</span></div>';
  }).join('');
  el.innerHTML = rows || '<div class="dd-none">—</div>';
}

async function farmMediaClear(kind) {
  const ips = farmMediaTargets();
  if (!ips.length) { toast(t('toastNoPrinters'), 'er'); return; }
  const label = kind==='video' ? t('mediaClearVideo') : t('mediaClearSnap');
  if (!confirm(label + '\n' + ips.length + ' ' + t('mediaFarmCount') + '\n\n' + t('mediaClearConfirm'))) return;
  const el = document.getElementById('sync-media-result');
  if (el) el.innerHTML = '<div class="dd-none">'+t('mediaLoading')+'</div>';
  try {
    const r = await fetch('/api/media/clear-all?kind='+kind, {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ips})});
    const d = await r.json();
    farmMediaResult(d.results);
    const total = (d.results||[]).reduce((n,x)=>n+(x.deleted||0),0);
    toast(label + ': ' + total, 'ok');
  } catch(e) { toast(e.message, 'er'); }
}

async function farmTimelapse(on, enforce) {
  const ips = farmMediaTargets();
  if (!ips.length) { toast(t('toastNoPrinters'), 'er'); return; }
  const el = document.getElementById('sync-media-result');
  if (el) el.innerHTML = '<div class="dd-none">'+t('mediaLoading')+'</div>';
  try {
    const r = await fetch('/api/media/timelapse-all?on='+(on?'1':'0')+'&enforce='+(enforce?'1':'0'),
      {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ips})});
    const d = await r.json();
    farmMediaResult(d.results);
    // Lokale Enforce-Anzeige aktualisieren.
    if (!MEDIA_ENFORCE) MEDIA_ENFORCE = {};
    ips.forEach(ip => { if (enforce && !on) MEDIA_ENFORCE[ip] = true; else delete MEDIA_ENFORCE[ip]; });
    toast(t('mediaTimelapse') + ': ' + (on ? t('on') : t('off')), 'ok');
  } catch(e) { toast(e.message, 'er'); }
}

// buildXcamControls: make the built-in AI monitoring and the
// abort sensitivity (Halt sensitivity) settable right here (M3).
// Best-effort — not every firmware reacts; the display follows the reported
// Zustand des Druckers.
function buildXcamControls(p) {
  const s = STATUS[p.ip];
  if (!s || !s.online) return '';
  const hatMon = typeof s.ai_monitoring !== 'undefined';
  const hatSens = !!(s.xcam && s.xcam.halt_print_sensitivity);
  if (!hatMon && !hatSens) return '';
  let h = '<div class="dd-sep"></div>';
  if (hatMon) {
    const on = !!s.ai_monitoring;
    h += '<div class="dd-row"><span class="dd-lbl">' + t('aiMonitor') + '</span>'
      + '<button class="hbtn btn-xs ' + (on ? 'green' : '') + '" onclick="toggleXcam(\'' + esc(p.ip) + '\',' + (!on) + ')">'
      + (on ? t('aiOn') : t('aiOff')) + '</button></div>';
  }
  const cur = hatSens ? String(s.xcam.halt_print_sensitivity) : '';
  const opts = ['low', 'medium', 'high'];
  let sel = '<select class="dd-sel" onchange="setXcamSens(\'' + esc(p.ip) + '\',this.value)">';
  if (!cur || opts.indexOf(cur) < 0) sel += '<option value="" selected>' + esc(cur || '—') + '</option>';
  opts.forEach(o => { sel += '<option value="' + o + '"' + (o === cur ? ' selected' : '') + '>' + t('sens_' + o) + '</option>'; });
  sel += '</select>';
  h += '<div class="dd-row"><span class="dd-lbl">' + t('aiSensitivity') + '</span>' + sel + '</div>';
  return h;
}

async function toggleXcam(ip, on) {
  try {
    const r = await fetch('/api/xcam', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ip, monitoring: on})});
    if (!r.ok) throw new Error('HTTP ' + r.status);
    if (STATUS[ip]) STATUS[ip].ai_monitoring = on;
    ddRefresh(ip);
    toast(t('xcamSetMsg'), 'ok');
  } catch(e) { toast('Fehlgeschlagen', 'er'); }
}

async function setXcamSens(ip, val) {
  if (!val) return;
  try {
    const r = await fetch('/api/xcam', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ip, sensitivity: val})});
    if (!r.ok) throw new Error('HTTP ' + r.status);
    if (STATUS[ip]) { if (!STATUS[ip].xcam) STATUS[ip].xcam = {}; STATUS[ip].xcam.halt_print_sensitivity = val; }
    ddRefresh(ip);
    toast(t('xcamSetMsg'), 'ok');
  } catch(e) { toast('Fehlgeschlagen', 'er'); }
}

// ─── AMS UND FILAMENT ─────────────────────────────────────────────────────────

// Colors come as RRGGBBAA. The AA is discarded, transparent
// filaments do not exist and the alpha channel is usually FF anyway.
function trayColorCSS(hex) {
  if (!hex || hex.length < 6) return '';
  return '#' + hex.substring(0, 6);
}

// trayDot draws the color dot. Multi-color filament yields several values
// in "cols" — these are shown as a gradient.
function trayDot(tr) {
  const cols = (tr.colors || []).map(trayColorCSS).filter(Boolean);
  if (cols.length > 1) {
    return '<span class="ams-dot multi" style="--grad:linear-gradient(135deg,' + cols.join(',') + ')"></span>';
  }
  const col = cols[0] || trayColorCSS(tr.color);
  return '<span class="ams-dot' + (col ? '' : ' empty') + '"' + (col ? ' style="--tray:' + col + '"' : '') + '></span>';
}

function trayLabel(tr) {
  const name = tr.sub_brand || tr.type;
  if (!name) return 'leer';
  return name;
}

// deviceRows adds operating time and firmware.
//
// About operating time, honestly: the printer does not expose its own counter over
// the local interface — neither the status message nor
// get_version contains it. What is shown here is counted by this program itself,
// since the printer was added. So it says "measured since added"
// there; it would be dishonest to present it as a device counter.
// amsName turns "ams/0" into "AMS 1" — the device counts from
// zero, the user counts from one.
// amsPunkt zeichnet denselben Farbpunkt wie im Grid — einfarbig, mehrfarbig
// als Verlauf, leer als Schraffur.
function amsPunkt(tr) {
  const cols = (tr && tr.colors) || [];
  if (cols.length > 1) {
    return '<span class="ams-dot multi" style="--grad:linear-gradient(135deg,'
      + cols.map(c => '#' + esc(String(c).slice(0, 6))).join(',') + ')"></span>';
  }
  const col = tr && tr.color ? '#' + esc(String(tr.color).slice(0, 6)) : '';
  return '<span class="ams-dot' + (col ? '' : ' empty') + '"'
    + (col ? ' style="--tray:' + col + '"' : '') + '></span>';
}

function amsName(modul) {
  const n = String(modul || '').toLowerCase().replace(/^ams[\/_]?/, '');
  if (n === '') return 'AMS';
  const z = parseInt(n, 10);
  return isNaN(z) ? 'AMS ' + n.toUpperCase() : 'AMS ' + (z + 1);
}

// fwPillHTML builds the "⬆ Update" badge for a printer — empty when nothing
// is pending. Used in the tile/list and in "Saved printers".
function fwPillHTML(ip) {
  const up = (STATUS[ip] && STATUS[ip].fw_updates) || [];
  if (!up.length) return '';
  const tip = up.map(u => fwModulName(u.modul) + ': ' + (u.aktuell || '?') + ' → ' + u.neu).join('\n');
  return '<span class="fw-pill" title="' + esc(tip) + '" onclick="event.stopPropagation()">⬆ Update</span>';
}

// fwModulName turns the module into a readable label.
function fwModulName(modul) {
  const m = String(modul || '').toLowerCase();
  if (m === 'ota') return 'Firmware';
  if (m.startsWith('ams')) return amsName(m);
  return modul || 'Modul';
}

// searchPrinterUpdates asks the server to freshly poll each printer. The
// server triggers no update — it only reads what the printer reports.
async function searchPrinterUpdates() {
  const btn = document.getElementById('fw-btn');
  const st = document.getElementById('fw-status');
  if (!btn) return;
  const alt = btn.textContent;
  btn.disabled = true;
  btn.textContent = 'Suche läuft …';
  if (st) st.textContent = '';
  try {
    const r = await fetch('/api/fwupdate/scan', { method: 'POST' });
    const d = await r.json();
    if (st) {
      st.textContent = (d.mit_update > 0)
        ? '✓ ' + d.mit_update + ' Drucker mit Update — ' + d.geprueft + ' geprüft'
        : '✓ Keine Updates offen (' + d.geprueft + ' von ' + (d.angefragt || 0) + ' erreichbar geprüft)';
    }
    await loadStatus();   // Pillen neu zeichnen
    renderQL();           // "Gespeicherte Drucker" mit Marken
    renderGrid();
    showFwList();
    showFwRaw(d.roh);
  } catch (e) {
    if (st) st.textContent = '✗ Fehler bei der Suche';
  } finally {
    btn.disabled = false;
    btn.textContent = alt;
  }
}

// showFwRaw exposes the raw upgrade_state messages. This helps when a
// model (e.g. X1E) structures its update info differently than expected — then
// one can see here what the printer really sends.
// showFwList lists the printers with a pending update plus "Start update".
function showFwList() {
  const box = document.getElementById('fw-liste');
  if (!box) return;
  const mit = P.filter(p => (STATUS[p.ip] && STATUS[p.ip].fw_updates || []).length);
  if (!mit.length) { box.innerHTML = ''; return; }
  box.innerHTML = mit.map(p => {
    const ups = STATUS[p.ip].fw_updates || [];
    const detail = ups.map(u => fwModulName(u.modul) + ': ' + (u.aktuell || '?') + ' → ' + u.neu).join(', ');
    return '<div class="fw-item">'
      + '<div class="fw-item-info"><span class="fw-item-name">' + esc(p.name) + '</span>'
      + '<span class="fw-item-detail">' + esc(detail) + '</span></div>'
      + '<button class="hbtn btn-sm red" onclick="startPrinterUpdate(\'' + esc(p.ip) + '\')">'+t('fwBtnStart')+'</button>'
      + '</div>';
  }).join('');
}

// startPrinterUpdate warns clearly and then triggers the update.
async function startPrinterUpdate(ip) {
  const p = P.find(x => x.ip === ip);
  if (!p) return;
  if (!confirm(t('fwWarn').replace('{name}', p.name))) return;
  try {
    const r = await fetch('/api/fwupdate/start', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ip})});
    const d = await r.json();
    if (r.ok && d.ok) toast('Update-Befehl an ' + p.name + ' gesendet', 'ok');
    else toast('Fehlgeschlagen: ' + (d.error || 'unbekannt'), 'er');
  } catch(e) { toast('Fehlgeschlagen: ' + (e.message || e), 'er'); }
}

function showFwRaw(roh) {
  const box = document.getElementById('fw-roh');
  if (!box) return;
  const eintraege = Object.keys(roh || {});
  if (!eintraege.length) { box.innerHTML = ''; return; }
  let inner = '';
  eintraege.forEach(ip => {
    const p = P.find(x => x.ip === ip);
    inner += '<div><b>' + esc(p ? p.name : ip) + '</b> (' + esc(ip) + ')\n' + esc(roh[ip]) + '</div>\n\n';
  });
  box.innerHTML = '<details><summary>Rohmeldung der Drucker (upgrade_state) — zum Melden bei fehlender Anzeige</summary><pre>'
    + inner + '</pre></details>';
}

function deviceRows(p) {
  const s = STATUS[p.ip] || {};
  let h = '';
  if (s.laufzeit_text) {
    h += '<div class="dd-row"><span class="dd-lbl">⏳ Laufzeit</span>'
       + '<span class="dd-val" title="von diesem Programm mitgezählt, nicht der Zähler im Gerät">'
       + esc(s.laufzeit_text) + ' <span class="dd-fein">gemessen seit Eintrag</span></span></div>';
  }
  const info = s.info;
  if (info && info.firmware) {
    h += '<div class="dd-row"><span class="dd-lbl">⚙ Firmware</span><span class="dd-val mono">'
       + esc(info.firmware) + '</span></div>';
  }
  return h;
}

function buildAMSBlock(p) {
  const s = STATUS[p.ip];
  if (!s || !s.online) return '';
  const units = s.ams || [];
  const ext = s.ext_spool;
  if (!units.length && !(ext && ext.type)) return '';

  let html = '<div class="dd-sep"></div>'
    + '<div class="dd-row"><span class="dd-lbl">AMS</span><span class="dd-val">'
    + (units.length ? units.length + (units.length === 1 ? ' Einheit' : ' Einheiten') : 'nur externe Spule')
    + '</span></div>';

  // Firmware of the AMS units from the device's module list.
  const amsInfo = (s.info && s.info.ams) || [];
  amsInfo.forEach(m => {
    html += '<div class="dd-row"><span class="dd-lbl">└ ' + esc(amsName(m.name)) + '</span>'
         + '<span class="dd-val mono">' + esc(m.sw_ver || '–')
         + (m.hw_ver ? ' <span class="dd-fein">' + esc(m.hw_ver) + '</span>' : '')
         + '</span></div>';
  });

  units.forEach((u, ui) => {
    html += '<div class="ams-unit"><div class="ams-head">AMS ' + (parseInt(u.id) + 1 || ui + 1)
      + (u.humidity ? ' · Feuchte ' + esc(u.humidity) : '')
      + (u.temp && parseFloat(u.temp) > 0 ? ' · ' + esc(u.temp) + '°' : '') + '</div>'
      + '<div class="ams-trays">';
    (u.trays || []).forEach(tr => {
      const col = trayColorCSS(tr.color);
      // tray_now counts across all AMS: unit * 4 + tray
      const globalSlot = (parseInt(u.id) || 0) * 4 + (parseInt(tr.slot) || 0);
      const active = String(globalSlot) === String(s.tray_now);
      const belegt = !!(col || (tr.colors && tr.colors.length) || tr.type);
      const geladen = belegt && !active;
      html += '<span class="ams-tray' + (active ? ' active' : (geladen ? ' geladen' : '')) + '" title="' + esc(trayLabel(tr))
        + (tr.remain > 0 ? ' · ' + tr.remain + '% Rest' : '') + (active ? ' · ' + t('amsActive') : (geladen ? ' · ' + t('amsLoaded') : '')) + '">'
        + (active ? '<span class="ams-active-mark">▶</span>' : '') + trayDot(tr) + esc(trayLabel(tr)) + '</span>';
    });
    html += '</div></div>';
  });

  if (ext && ext.type) {
    html += '<div class="ams-unit"><div class="ams-head">Externe Spule</div><div class="ams-trays">'
      + '<span class="ams-tray">' + trayDot(ext) + esc(trayLabel(ext)) + '</span>'
      + '</div></div>';
  }
  return html;
}

// ─── DRUCKSTEUERUNG ───────────────────────────────────────────────────────────

function buildPrintControls(p) {
  const s = STATUS[p.ip];
  if (!p.serial) return '';
  const online = !!(s && s.online);
  const state = online ? (s.gcode_state || '').toUpperCase() : '';
  const running = state === 'RUNNING' || state === 'PREPARE' || state === 'SLICING';
  const paused = state === 'PAUSE';
  const active = running || paused;
  const ip = esc(p.ip);

  return '<div class="dd-sep"></div><div class="pc-row">'
    + '<button class="pc-btn" ' + (running ? '' : 'disabled ') + 'onclick="printCmd(\'' + ip + '\',\'pause\')">'+t('printPause')+'</button>'
    + '<button class="pc-btn" ' + (paused ? '' : 'disabled ') + 'onclick="printCmd(\'' + ip + '\',\'resume\')">'+t('printResume')+'</button>'
    + '<button class="pc-btn danger" ' + (active ? '' : 'disabled ') + 'onclick="printCmd(\'' + ip + '\',\'stop\')">'+t('printCancel')+'</button>'
    + '</div>'
    + '<div class="pc-row"><button class="pc-btn" ' + (active ? 'disabled ' : '')
      + 'onclick="openPrintStart(\'' + ip + '\')">'+t('printSelectFile')+'</button>'
      + '<input type="file" id="mup-' + ip + '" accept=".gcode,.3mf" style="display:none" onchange="modalUpload(\'' + ip + '\', this)">'
      + '<button class="pc-btn" ' + (active ? 'disabled ' : '') + 'onclick="byId(\'mup-' + ip + '\').click()">'+t('modalUpload')+'</button></div>'
    + '<div class="ds-box hidden" id="ds-' + ip + '"></div>'
    + (online ? '' : '<div class="dd-none mt5">Steuerung braucht eine MQTT-Verbindung</div>');
}

// modalUpload uploads a file directly from the modal to THIS printer (FTPS).
// Uses the same endpoint as File Sync, just for one IP.
async function modalUpload(ip, input){
  const f = input.files && input.files[0];
  input.value = '';
  if (!f) return;
  const nm = f.name.toLowerCase();
  if (!nm.endsWith('.gcode') && !nm.endsWith('.3mf')) { toast(t('uploadBadFormat'), 'er'); return; }
  toast(t('uploadStart') + ' ' + f.name, 'ok');
  try {
    const fd = new FormData(); fd.append('file', f);
    const r = await fetch('/api/sync/upload?ip=' + encodeURIComponent(ip), {method:'POST', body: fd});
    if (!r.ok) { let m=''; try { m = (await r.json()).error; } catch(e){} throw new Error(m || ('HTTP ' + r.status)); }
    toast(t('uploadDone') + ' ' + f.name, 'ok');
  } catch(e) { toast(e.message || String(e), 'er'); }
}

// ─── TEMPERATUR ───────────────────────────────────────────────────────────────
//
// Control nozzle and bed individually: minus/plus in 5-degree steps, a field to
// type in and a reset button that resets to the target reported by the printer.
// M104/M140 is sent, without waiting.
function buildTempControls(p) {
  const s = STATUS[p.ip];
  if (!s || !s.online) return '';
  const ip = esc(p.ip);
  const nz = Math.round(s.nozzle_target || 0), bd = Math.round(s.bed_target || 0);
  const zeile = (was, label, ist, soll) =>
    '<div class="tc-zeile"><span class="tc-lbl">' + label + '</span>'
    + '<span class="tc-ist">' + Math.round(ist || 0) + '°</span>'
    + '<button class="tc-btn" data-ip="' + ip + '" data-w="' + was + '" onclick="event.stopPropagation();tempStep(this.dataset.ip,this.dataset.w,-1)">−</button>'
    + '<input class="tc-feld" id="tc-' + was + '-' + ip + '" type="number" value="' + soll + '" min="0">'
    + '<button class="tc-btn" data-ip="' + ip + '" data-w="' + was + '" onclick="event.stopPropagation();tempStep(this.dataset.ip,this.dataset.w,1)">+</button>'
    + '<button class="tc-btn tc-set" data-ip="' + ip + '" data-w="' + was + '" onclick="event.stopPropagation();tempSet(this.dataset.ip,this.dataset.w)">'+t('setBtn')+'</button>'
    + '<button class="tc-btn" data-ip="' + ip + '" data-w="' + was + '" title="zurück auf den gemeldeten Sollwert" onclick="event.stopPropagation();tempReset(this.dataset.ip,this.dataset.w)">↺</button>'
    + '</div>';
  return '<div class="dd-sep"></div><div class="tc-block">'
    + zeile('nozzle', 'Düse', s.nozzle_temp, nz)
    + zeile('bed', 'Bett', s.bed_temp, bd)
    + '</div>';
}

function tempStep(ip, was, delta) {
  const f = byId('tc-' + was + '-' + ip);
  if (!f) return;
  const v = Math.max(0, (parseInt(f.value, 10) || 0) + delta);
  f.value = v;
}

function tempReset(ip, was) {
  const s = STATUS[ip] || {};
  const f = byId('tc-' + was + '-' + ip);
  if (f) f.value = Math.round((was === 'nozzle' ? s.nozzle_target : s.bed_target) || 0);
}

async function tempSet(ip, was) {
  const f = byId('tc-' + was + '-' + ip);
  if (!f) return;
  const temp = parseInt(f.value, 10);
  if (isNaN(temp)) { toast('Bitte eine Zahl eintragen', 'er'); return; }
  busyStart('setzt Temperatur …');
  try {
    const r = await fetch('/api/print/temp', {method:'POST', headers:{'Content-Type':'application/json'},
      body: JSON.stringify({ip, was, temp})});
    const d = await r.json().catch(() => ({}));
    if (!r.ok) throw new Error(d.error || (await r.text()) || ('HTTP ' + r.status));
    if (d.ok === false) throw new Error(d.error || 'abgelehnt');
    toast((was === 'nozzle' ? 'Düse' : 'Bett') + ' auf ' + temp + '° gesetzt', 'ok');
  } catch(e) { toast('Fehlgeschlagen: ' + (e.message || e), 'er'); }
  finally { busyEnd(); }
}

// ─── AMS-FACH BEARBEITEN ───────────────────────────────────────────────────────
//
// Vendor, material and color per tray directly in the AMS overview. The
// printer requires the values as a profile id (tray_info_idx) that
// sets vendor AND material at once — so the matching id is looked up
// from the selection. It does not accept a free vendor field.
const AMS_HERSTELLER = ['Standard', 'Generic', 'PolyLite', 'Andere'];
const AMS_MATERIAL   = ['PLA', 'PETG', 'ABS', 'PC', 'TPU', 'PA', 'ASA'];
const AMS_TEMP = {PLA:[190,230], PETG:[220,260], ABS:[240,270], PC:[260,280], TPU:[210,240], PA:[250,290], ASA:[240,270]};
// Known ids; where none exists, the mapping goes by type and color
// alone (empty id), which the printer also accepts.
const AMS_IDX = {
  'Standard|PLA':'GFA00','Standard|PETG':'GFG00','Standard|ABS':'GFB00','Standard|PC':'GFC00',
  'Generic|PLA':'GFL99','Generic|PETG':'GFG99','Generic|ABS':'GFB99','Generic|PC':'GFC99','Generic|PA':'GFN99',
  'PolyLite|PLA':'GFL00',
};

function amsIdxFuer(hersteller, material) {
  return AMS_IDX[hersteller + '|' + material] || '';
}

// amsSlotEditor renders the selection for a tray. amsUi/ti are AMS and
// tray index, tr the current values.
function amsSlotEditor(ip, amsUi, ti, tr) {
  const eip = esc(ip);
  const key = amsUi + '_' + ti;
  const typ = (tr && tr.type ? String(tr.type).toUpperCase() : 'PLA');
  const farbe = tr && tr.color ? '#' + String(tr.color).slice(0, 6) : '#22aaff';
  const opt = (arr, sel) => arr.map(x => '<option' + (x === sel ? ' selected' : '') + '>' + esc(x) + '</option>').join('');
  return '<div class="af-editor" id="af-' + eip + '-' + key + '">'
    + '<select class="af-sel" id="af-h-' + eip + '-' + key + '" title="Hersteller">' + opt(AMS_HERSTELLER, 'Standard') + '</select>'
    + '<select class="af-sel" id="af-m-' + eip + '-' + key + '" title="Material">' + opt(AMS_MATERIAL, AMS_MATERIAL.indexOf(typ) >= 0 ? typ : 'PLA') + '</select>'
    + '<input type="color" class="af-farbe" id="af-c-' + eip + '-' + key + '" value="' + esc(farbe) + '" title="Farbe">'
    + '<button class="af-set" data-ip="' + eip + '" data-a="' + amsUi + '" data-t="' + ti + '" onclick="event.stopPropagation();amsSlotSet(this.dataset.ip,this.dataset.a,this.dataset.t)" title="'+t('setBtn')+'">✓</button>'
    + '</div>';
}

async function amsSlotSet(ip, amsUi, ti) {
  const key = amsUi + '_' + ti;
  const val = pre => (byId('af-' + pre + '-' + ip + '-' + key) || {}).value || '';
  const hersteller = val('h'), material = (val('m') || 'PLA').toUpperCase();
  const temp = AMS_TEMP[material] || [190, 230];
  const body = {
    ip, ams_id: parseInt(amsUi, 10), tray_id: parseInt(ti, 10),
    tray_type: material,
    tray_color: (val('c') || '#000000').replace('#', ''),
    tray_info_idx: amsIdxFuer(hersteller, material),
    nozzle_temp_min: temp[0], nozzle_temp_max: temp[1],
  };
  busyStart('setzt Filament …');
  try {
    const r = await fetch('/api/ams/filament', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify(body)});
    const d = await r.json().catch(() => ({}));
    if (!r.ok) throw new Error(d.error || (await r.text()) || ('HTTP ' + r.status));
    if (d.ok === false) throw new Error(d.error || 'abgelehnt');
    toast('AMS ' + (parseInt(amsUi,10)+1) + ' Fach ' + (parseInt(ti,10)+1) + ' gesetzt', 'ok');
    setTimeout(loadStatus, 1200);
  } catch(e) { toast('Fehlgeschlagen: ' + (e.message || e), 'er'); }
  finally { busyEnd(); }
}

// buildAMSUnits shows all AMS with all four trays and an editor per
// tray. Used in the list (AMS box) and in the grid dropdown.
function buildAMSUnits(p) {
  const s = STATUS[p.ip];
  if (!s || !s.online) return '';
  const einheiten = (s.ams || []);
  const amsInfo = (s.info && s.info.ams) || [];
  const anzahl = Math.min(4, Math.max(einheiten.length, amsInfo.length));
  if (!anzahl && !(s.ext_spool && s.ext_spool.type)) return '';
  let html = '';
  for (let ui = 0; ui < anzahl; ui++) {
    const u = einheiten[ui] || {};
    const m = amsInfo[ui] || {};
    html += '<div class="af-unit"><div class="af-kopf">' + esc(amsName(m.name || ('ams/' + ui)))
      + (m.sw_ver ? ' <span class="dd-fein">' + esc(m.sw_ver) + (m.hw_ver ? ' · ' + esc(m.hw_ver) : '') + '</span>' : '') + '</div>';
    for (let ti = 0; ti < 4; ti++) {
      const tr = (u.trays || [])[ti] || {};
      html += '<div class="af-fach">'
        + '<span class="af-nr">' + (ti + 1) + '</span>'
        + amsPunkt(tr)
        + '<span class="af-typ">' + esc(tr.type || '–') + '</span>'
        + amsSlotEditor(p.ip, ui, ti, tr)
        + '</div>';
    }
    html += '</div>';
  }
  if (s.ext_spool && s.ext_spool.type) {
    html += '<div class="af-unit"><div class="af-kopf">Externe Rolle</div>'
      + '<div class="af-fach"><span class="af-nr">·</span>' + amsPunkt(s.ext_spool)
      + '<span class="af-typ">' + esc(s.ext_spool.type) + '</span></div></div>';
  }
  return html;
}

const PRINT_CMD_LABEL = {pause: 'pausiert', resume: 'fortgesetzt', stop: 'abgebrochen'};

async function printCmd(ip, cmd) {
  const p = P.find(x => x.ip === ip);
  const name = p ? p.name : ip;
  // Cancelling costs the running print — so a confirmation with the name,
  // pause and resume go out directly.
  if (cmd === 'stop' && !confirm('Druck auf "' + name + '" wirklich abbrechen?\n\nDas beendet den laufenden Auftrag endgültig.')) return;

  try {
    const r = await fetch('/api/print/command', {
      method: 'POST', headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({ips: [ip], command: cmd})
    });
    const d = await r.json();
    if (!r.ok) throw new Error(d && d.error ? d.error : ('HTTP ' + r.status));
    if (d.sent > 0) {
      toast(name + ' ' + PRINT_CMD_LABEL[cmd], 'ok');
      setTimeout(loadStatus, 1500);
    } else {
      const why = (d.failed && d.failed[0] && d.failed[0].error) || 'kein Drucker erreicht';
      toast('Fehlgeschlagen: ' + why, 'er');
    }
  } catch(e) {
    toast('Fehlgeschlagen: ' + (e.message || e), 'er');
  }
}



function toggleUebersicht(e) {
  e.stopPropagation();
  const menu = document.getElementById('uebersicht-menu');
  setShown(menu, menu.classList.contains('hidden'));
}
function closeUebersicht() {
  const menu = document.getElementById('uebersicht-menu');
  if (menu) setShown(menu, false);
}
document.addEventListener('click', () => closeUebersicht());
let DD_OPEN_IP = null;
let DD_MODAL = false;

function closeDD() {
  const pop = document.getElementById('dd-pop');
  if (pop) { pop.classList.remove('open', 'dd-modal'); pop.style.left = ''; pop.style.top = ''; pop.innerHTML = ''; }
  const bd = document.getElementById('dd-backdrop'); if (bd) bd.classList.remove('open');
  document.querySelectorAll('.tile-name-arrow.open').forEach(el => el.classList.remove('open'));
  DD_OPEN_IP = null;
  DD_MODAL = false;
}

function toggleDD(ip) {
  if (DD_OPEN_IP === ip) { closeDD(); return; }
  closeDD();
  const anchor = document.querySelector('.tile[data-ip="' + CSS.escape(ip) + '"] .tile-name');
  const pop = document.getElementById('dd-pop');
  if (!anchor || !pop) return;

  const p = P.find(x => x.ip === ip) || {ip};
  pop.innerHTML = '<div class="dd-title">' + esc(p.name || ip) + '<span class="dd-close" title="' + esc(t('close')) + '" onclick="event.stopPropagation();closeDD()">✕</span></div>' + buildDropdownContent(p);
  DD_OPEN_IP = ip;
  // In the grid view as a centered modal box (with backdrop), otherwise as
  // Dropdown am Kachelnamen.
  DD_MODAL = (typeof CURRENT_SUBVIEW !== 'undefined' && CURRENT_SUBVIEW === 'grid');
  const bd = document.getElementById('dd-backdrop');
  if (DD_MODAL) {
    pop.style.left = ''; pop.style.top = '';
    pop.classList.add('open', 'dd-modal');
    if (bd) bd.classList.add('open');
  } else {
    pop.classList.remove('dd-modal');
    if (bd) bd.classList.remove('open');
    pop.classList.add('open');
    positionDD(anchor, pop);
  }

  const arr = document.getElementById('arr-' + ip);
  if (arr) arr.classList.add('open');
}

function positionDD(anchor, pop) {
  const r = anchor.getBoundingClientRect();
  const w = pop.offsetWidth, h = pop.offsetHeight;
  let left = r.left;
  if (left + w > window.innerWidth - 8) left = window.innerWidth - w - 8;
  if (left < 8) left = 8;
  let top = r.bottom + 4;
  if (top + h > window.innerHeight - 8) top = Math.max(8, r.top - h - 4);
  pop.style.left = Math.round(left) + 'px';
  pop.style.top = Math.round(top) + 'px';
}

document.addEventListener('click', e => {
  if (e.target.closest('#dd-pop')) return;      // Klicks im Menue nicht closeSplashInner
  if (!e.target.closest('.tile-name-wrap')) closeDD();
});
window.addEventListener('resize', () => { if (!DD_MODAL) closeDD(); });
window.addEventListener('scroll', () => { if (!DD_MODAL) closeDD(); }, true);

// The bar is ALWAYS created, even without data. Previously it only appeared with
// the first status message — a freshly added printer thus looked like
// a broken one, and the layout jumped as soon as data arrived.
function buildStatusBar(p) {
  const s = STATUS[p.ip];
  const state = s ? (s.gcode_state||'').toLowerCase() : '';
  const printing = !!(s && s.online && state && state !== 'idle');
  const pct = printing ? (s.progress || 0) : 0;
  const fillClass = {running:'running',pause:'paused',failed:'failed',finish:'finish'}[state]||'';
  return '<div class="ts-bar"><div class="ts-fill '+fillClass+'" style="--w:'+pct+'%"></div></div>'
       + '<div class="tile-meta" id="tmeta-'+esc(p.ip)+'">'+statusMetaText(p)+'</div>';
}

// statusMetaText is the line under the bar: progress, remaining time, layer
// — the same values that were previously only in the list view.
function statusMetaText(p) {
  const s = STATUS[p.ip];
  if (!p.serial) return '<span class="tm-wait">keine Seriennummer</span>';
  if (!s || !s.online) return '<span class="tm-wait">warte auf Daten …</span>';
  const state = (s.gcode_state||'').toLowerCase();
  if (!state || state === 'idle') return '<span class="tm-wait">bereit</span>';
  const parts = [];
  if (typeof s.progress === 'number') parts.push(s.progress + '%');
  if (s.remain_time > 0) parts.push('⏱ ' + fmtRemaining(s.remain_time * 60));
  return parts.length ? esc(parts.join('  ·  ')) : '<span class="tm-wait">läuft …</span>';
}

// Returns state pill HTML for the tile header
function buildStatePill(p) {
  const s = STATUS[p.ip];
  if (!p.serial || !s || !s.online) return '';
  const state = (s.gcode_state||'').toLowerCase();
  if (!state || state === 'idle') return '';
  const labels = {running:'PRINTING', pause:'PAUSED', failed:'FAILED', finish:'DONE'};
  const label = labels[state] || state.toUpperCase();
  return '<span class="tile-state '+state+'">'+label+'</span>';
}


// Update only the status bar of a tile — never touches the iframe/stream
// ─── EINE AKTUALISIERUNG FUER BEIDE ANSICHTEN ────────────────────────────────
//
// Previously two functions sat side by side here: updateStatusBar updated the
// tile, updateListRow the row — with their own error detection, their own
// percentage math, its own state list. Two truths for the same thing.
// Now there is one: the fields carry the same classes in both views,
// also genuegt es, sie aus printerFields() neu zu erzeugen.
const DF_KLASSE = {stern:'df-stern', modell:'df-modell', name:'df-name', datei:'df-datei',
                   fortschritt:'df-fortschritt', restzeit:'df-rest', ip:'df-ip', status:'df-status',
                   online:'df-online-wrap'};

let CAM_KEY = {}; // zuletzt gezeigter Kamera-Zustand je Drucker
function refreshTile(ip) {
  const p = P.find(x => x.ip === ip);
  if (!p) return;
  const f = printerFields(p);

  // The camera area strictly follows the KNOWN state of the
  // printer — not on whether a video stream comes up. From the
  // status we know online/offline/in repair; exactly that we show immediately.
  // buildTile set CAM_KEY when drawing; if the state changes, the
  // tile is rebuilt (and any running spinner/stream stopped).
  const camKeyNeu = isRepair(ip) ? 'repair' : (isCameraOff(ip) ? 'private' : (!f.online ? 'offline' : 'online'));
  const tileDa = document.querySelector('.tile[data-ip="' + CSS.escape(ip) + '"]');
  if (tileDa && CAM_KEY[ip] !== camKeyNeu) {
    if (camKeyNeu !== 'online') { stopWebRTC(ip); stopSnapPoll(ip); } // Stream + Ladekreis weg
    renderTile(ip); // setzt CAM_KEY[ip] neu
    if (camKeyNeu === 'online' && (ST[ip] === 'live' || ST[ip] === 'snap')) mountStreams([p]);
  }
  // The red error border can change without a camera-state change (a print
  // fails, or the error is acknowledged) — keep the grid tile in sync.
  { const tl = document.querySelector('.tile[data-ip="' + CSS.escape(ip) + '"]');
    if (tl) tl.classList.toggle('has-error', f.hatFehler); }

  document.querySelectorAll('.df-item[data-ip="' + CSS.escape(ip) + '"]').forEach(el => {
    el.classList.toggle('fav', f.fav);
    el.classList.toggle('hat-fehler', f.hatFehler);
    el.classList.toggle('has-error', f.hatFehler);
    el.classList.toggle('pausiert', f.pausiert);
    el.classList.toggle('fertig', f.fertig);
    Object.keys(DF_KLASSE).forEach(n => {
      const alt = el.querySelector('.' + DF_KLASSE[n]);
      if (alt) alt.outerHTML = feldHTML(f, n);
    });
  });

  // Carry the open detail row along — but NOT while someone is working in it.
  // Previously it was fully rebuilt on every status tick; an
  // open file picker or a half-filled form was
  // thus gone after five seconds.
  if (typeof LIST_AUF !== 'undefined' && LIST_AUF.has(ip) && detailFrei(ip)) drawListRow(ip);
  if (DD_OPEN_IP === ip) {
    const pop = document.getElementById('dd-pop');
    // Only redraw when nobody is working in the menu.
    if (pop && ddFrei()) pop.innerHTML = '<div class="dd-title">' + esc(f.name) + '<span class="dd-close" title="' + esc(t('close')) + '" onclick="event.stopPropagation();closeDD()">✕</span></div>' + buildDropdownContent(p);
  }
  updateFSVCell(ip, f); // Rahmenfarbe/Balken im „Fullscreen Video Only" mitziehen

  // Update the online pill in "Saved printers" too — otherwise it is only set on
  // draw (often before the first status) and permanently shows "offline".
  const qlrow = document.getElementById('qlrow-' + ip);
  if (qlrow) {
    const stz = qlrow.querySelector('.ql-status');
    if (stz) stz.innerHTML = onlinePill(ip); // nicht im Bearbeiten-Modus (kein .ql-status)
    // Carry the repair badge on the name along (e.g. when another PC sets it).
    const nm = qlrow.querySelector('.ql-name');
    if (nm) nm.innerHTML = '<span class="ql-name-txt">' + esc(f.name) + '</span>' + fwPillHTML(ip) + repairPill(ip);
  }
}

// ddFrei: do not rebuild while a form is open or the cursor is in
// einem Eingabefeld des Menues steht.
function ddFrei() {
  const pop = document.getElementById('dd-pop');
  if (!pop) return true;
  if (pop.querySelector('.ds-box:not(.hidden)')) return false;
  const a = document.activeElement;
  if (a && a !== document.body && pop.contains(a)) return false;
  return true;
}

// The old names remain as pass-throughs — they are called in many
// places, and renaming only brought risk without benefit.
function updateStatusBar(ip) { refreshTile(ip); }
function updateListRow(ip)   { refreshTile(ip); }

// ─── STANDALONE-KAMERAS (Frontend) ───────────────────────────────────────────
// Cameras are tied to no printer object. They are loaded separately, appended at the end
// of the tiles and carry NO printer overlays (no progress, no
// temperature, no AMS, no status color). WebRTC/snapshot run over the same
// mechanism as for printers, just with the camera ID as key.
let CAMS = [];
const CAM_IDS = new Set();
function isCamera(key){ return CAM_IDS.has(key); }
async function loadCameras(){
  try { const r = await fetch('/api/cameras', {cache:'no-store'}); CAMS = (await r.json()) || []; }
  catch(e){ CAMS = []; }
  CAM_IDS.clear(); CAMS.forEach(c => CAM_IDS.add(c.id));
}

// renderCamAdmin shows the configured cameras with a delete button. The source
// (with credentials) is deliberately shown masked.
function renderCamAdmin(){
  const el = document.getElementById('cam-list'); if(!el) return;
  if(!CAMS.length){ el.innerHTML = '<p class="note-sm">'+t('camNone')+'</p>'; return; }
  el.innerHTML = CAMS.map(c =>
    '<div class="cam-row"><span class="cam-row-name">'+esc(c.name)+'</span>'
    + '<span class="mono note-sm">'+esc(c.stream||c.id)+'</span>'
    + '<span class="mono note-sm cam-row-src">'+esc(maskSource(c.source))+'</span>'
    + '<button class="hbtn btn-sm red" onclick="delCamera(\''+esc(c.id)+'\')">'+t('camDelete')+'</button></div>'
  ).join('');
}
// maskSource shows the source without the password (everything after the first ':' up to '@').
function maskSource(src){
  if(!src) return '';
  return String(src).replace(/(\/\/[^:@/]+):[^@/]*@/, '$1:•••@');
}
async function addCamera(){
  const name=(document.getElementById('cam-name').value||'').trim();
  const source=(document.getElementById('cam-src').value||'').trim();
  if(!name || !source){ toast(t('camNeedFields'),'er'); return; }
  try{
    const r=await fetch('/api/cameras',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name,source})});
    if(!r.ok) throw new Error((await r.text()).trim()||('HTTP '+r.status));
    document.getElementById('cam-name').value=''; document.getElementById('cam-src').value='';
    await loadCameras(); render();
    toast(t('camAdded'),'ok');
  }catch(e){ toast((e.message||e),'er'); }
}
let XM_CHALLENGE = null; // 'captcha' | 'verify'
function xmResult(msg){ const o=document.getElementById('xm-result'); if(o) o.textContent=msg; }
function xmForm(obj){ return Object.keys(obj).map(k=>encodeURIComponent(k)+'='+encodeURIComponent(obj[k])).join('&'); }
// Login/confirmation ALWAYS as application/x-www-form-urlencoded — exactly what
// go2rtc's /api/xiaomi expects (r.ParseForm). Credentials are only in the body.
async function xmSend(fields){
  return fetch('/api/xiaomi-login',{method:'POST',headers:{'Content-Type':'application/x-www-form-urlencoded'},body:xmForm(fields)});
}
async function xiaomiLogin(){
  const username=(document.getElementById('xm-user').value||'').trim();
  const password=(document.getElementById('xm-pass').value||'');
  if(!username || !password){ xmResult(t('xmNeedCreds')); return; }
  xmResult(t('xmRunning'));
  try{
    const r=await xmSend({username,password});
    document.getElementById('xm-pass').value=''; // sofort leeren
    await xmAfterAuth(r);
  }catch(e){ xmResult(String(e.message||e)); }
}
async function xiaomiChallenge(){
  const code=(document.getElementById('xm-code').value||'').trim();
  if(!code){ return; }
  xmResult(t('xmRunning'));
  try{
    const fields = XM_CHALLENGE==='captcha' ? {captcha:code} : {verify:code};
    const r=await xmSend(fields);
    document.getElementById('xm-code').value='';
    await xmAfterAuth(r);
  }catch(e){ xmResult(String(e.message||e)); }
}
async function xmAfterAuth(r){
  const txt=await r.text();
  if(r.status===200){
    setShown(document.getElementById('xm-challenge'), false);
    XM_CHALLENGE=null;
    xmResult(t('xmOk'));
    toast(t('xmOk'),'ok');
    xiaomiListDevices();
    return;
  }
  if(r.status===401){
    let d={}; try{ d=JSON.parse(txt); }catch(e){}
    const ch=document.getElementById('xm-challenge');
    const cwrap=document.getElementById('xm-captcha-wrap');
    const hint=document.getElementById('xm-challenge-hint');
    if(d.captcha){
      XM_CHALLENGE='captcha';
      const mime = String(d.captcha).startsWith('iVBOR') ? 'image/png' : 'image/jpeg';
      document.getElementById('xm-captcha-img').src='data:'+mime+';base64,'+d.captcha;
      setShown(cwrap, true);
      if(hint) hint.textContent=t('xmCaptchaHint');
    } else if(d.verify_phone || d.verify_email){
      XM_CHALLENGE='verify';
      setShown(cwrap, false);
      if(hint) hint.textContent=t('xmVerifyHint').replace('%s', d.verify_phone||d.verify_email);
    } else {
      xmResult('HTTP 401\n'+txt); return;
    }
    setShown(ch, true);
    xmResult('');
    return;
  }
  xmResult('HTTP '+r.status+'\n'+txt);
}
// After a successful login: list accounts + cameras (GET, only ID+region in
// the URL — NO credentials). Each camera gets an "Add" button.
async function xiaomiListDevices(){
  // Cameras belong to a Mi region. The region chosen by the user first,
  // then try the other known regions — go2rtc returns 404 ("no sources") for a
  // region without cameras. So the camera is found even
  // when it is registered in a region other than "de".
  const primary=(document.getElementById('xm-server').value||'de').trim();
  const regions=[primary,'de','us','cn','i2','ru','sg',''].filter((v,i,a)=>a.indexOf(v)===i);
  const box=document.getElementById('xm-devices');
  if(box) box.innerHTML='<p class="note-sm">'+t('xmRunning')+'</p>';
  try{
    const ids=await (await fetch('/api/xiaomi-login',{cache:'no-store'})).json();
    if(!Array.isArray(ids) || !ids.length){ if(box) box.innerHTML='<p class="note-sm">'+t('xmNoAccount')+'</p>'; return; }
    const gesehen=new Set(); let html='';
    for(const id of ids){
      for(const region of regions){
        let resp;
        try{ resp=await fetch('/api/xiaomi-login?id='+encodeURIComponent(id)+'&region='+encodeURIComponent(region),{cache:'no-store'}); }catch(e){ continue; }
        if(!resp.ok) continue; // 404 = in dieser Region keine Kamera
        let data; try{ data=await resp.json(); }catch(e){ continue; }
        const srcs=(data && data.sources) || [];
        srcs.forEach(sq=>{
          if(!sq.url || gesehen.has(sq.url)) return;
          gesehen.add(sq.url);
          const reg = region || 'global';
          html+='<div class="cam-row"><span class="cam-row-name">'+esc(sq.name||'')+'</span>'
            +'<span class="mono note-sm cam-row-src">'+esc((sq.info||'')+'  ['+reg+']')+'</span>'
            +'<button class="hbtn btn-sm green" data-name="'+esc(sq.name||'')+'" data-url="'+esc(sq.url||'')+'" onclick="addCameraFrom(this)">'+t('camAddBtn')+'</button></div>';
        });
      }
    }
    if(box) box.innerHTML = html || '<p class="note-sm">'+t('xmNoCams')+'</p>';
  }catch(e){ if(box) box.innerHTML='<p class="note-sm">'+esc(String(e.message||e))+'</p>'; }
}
// addCameraFrom fills the camera fields from a found device and creates it.
async function addCameraFrom(btn){
  document.getElementById('cam-name').value = btn.dataset.name || 'Kamera';
  document.getElementById('cam-src').value = btn.dataset.url || '';
  await addCamera();
}

async function delCamera(id){
  try{
    const r=await fetch('/api/cameras/'+encodeURIComponent(id),{method:'DELETE'});
    if(!r.ok && r.status!==204) throw new Error('HTTP '+r.status);
    // Stop the camera's running connection/images so nothing keeps going.
    if(ST[id]==='live') stopWebRTC(id); if(ST[id]==='snap') stopSnapPoll(id); delete ST[id];
    await loadCameras(); render();
    toast(t('camRemoved'),'ok');
  }catch(e){ toast((e.message||e),'er'); }
}

// buildCamTile: a tile for a standalone camera. Same size/theming
// as printer tiles, but only name + image.
function buildCamTile(cam){
  const id = cam.id;
  const st = ST[id] || 'idle', isLive = st==='live', isSnap = st==='snap';
  const camAttrs = ' onclick="event.stopPropagation();tileCamClick(\''+esc(id)+'\')" class="clickable"';
  const g2startphase = !G2online && (G2_RESTARTING || (Date.now()-APP_START < 45000));
  let cam2 = '';
  if(!G2online && g2startphase) cam2 = '<div class="cam-wait"'+camAttrs+'><div class="cam-spinner"></div><div class="cip">'+t('camStarting')+'</div></div>';
  else if(!G2online) cam2 = '<div class="cam-idle"'+camAttrs+'><div class="cip">'+esc(cam.name)+'</div></div>';
  else if(isLive) cam2 = '<video class="tile-video" id="vid-'+esc(id)+'" autoplay muted playsinline'+camAttrs+'></video>'
    +'<div class="cam-connect" id="conn-'+esc(id)+'"><div class="cam-spinner"></div><div class="cip">'+t('camConnecting')+'</div></div>';
  else if(isSnap) cam2 = '<div class="snap-container"'+camAttrs+'><img id="snapimg-'+esc(id)+'" src="" class="fill-media" onerror="this.classList.add(\'hidden\');">'
    +'<div class="snap-ring" id="snapring-'+esc(id)+'"></div><div class="snap-stamp" id="snapts-'+esc(id)+'"></div></div>';
  else cam2 = '<div class="cam-idle ready"'+camAttrs+'><div class="ci">▶</div><div class="cip">'+t('camReady')+'</div></div>';
  const online = G2online; // Kamera gilt als erreichbar, wenn go2rtc läuft
  return '<div class="df-item kachel tile cam-tile '+(cam.fav?'fav ':'')+(isLive?'streaming ':'')
      +(currentViewProfile==='minimal'?'vp-minimal':'')+'" data-ip="'+esc(id)+'">'
    +'<div class="df-kopf">'
      +'<div class="tile-name-wrap"><div class="tile-name">'+esc(cam.name)+'</div></div>'
      +'<span class="df-stern'+(cam.fav?' on':'')+'" data-ip="'+esc(id)+'" onclick="event.stopPropagation();camToggleFav(this.dataset.ip)">'+(cam.fav?'★':'☆')+'</span>'
    +'</div>'
    +'<div class="df-pillen">'
      +'<span class="df-online '+(online?'on':'off')+'">'+(online?t('lblOnline'):t('lblOffline'))+'</span>'
      +'<span class="cam-pill">'+t('camBadge')+'</span>'
      +(isLive?'<span class="tile-live-badge'+(!G2online?' dim':'')+'">LIVE</span>':'')
    +'</div>'
    +'<div class="tile-cam cam-center">'+cam2+((isLive||isSnap)?'<button class="tile-fs" title="'+esc(t('videoFullscreen'))+'" onclick="event.stopPropagation();videoFullscreen(\''+esc(id)+'\')">⛶</button>':'')+'</div>'
    +'<div class="df-balken breit cam-spacer"><span class="df-balken-fuell"></span></div>'
    +'<div class="df-fuss"><div class="df-fuss-r1">&nbsp;</div><div class="df-fuss-r2">&nbsp;</div></div>'
    +'</div>';
}

// camToggleFav toggles the favorite marker of a standalone camera.
async function camToggleFav(id){
  const c = CAMS.find(x=>x.id===id); if(!c) return;
  const neu = !c.fav; c.fav = neu;
  render();
  try{ await fetch('/api/cameras/'+encodeURIComponent(id), {method:'PATCH', headers:{'Content-Type':'application/json'}, body: JSON.stringify({fav:neu})}); }
  catch(e){ c.fav=!neu; render(); }
}

// mountCamStreams supplies the camera tiles — like mountStreams, but with the
// go2rtc stream name from the camera config instead of a printer name.
function mountCamStreams(){
  let started = 0;
  CAMS.forEach(cam => {
    const id = cam.id, st = ST[id];
    if(st === 'live'){
      if(PCS[id] && attachVideo(id)) return;
      const delay = 120 + (started++) * 80;
      setTimeout(() => { if(ST[id] === 'live') startWebRTC(id, cam.stream || id); }, delay);
    } else if(st === 'snap'){
      startSnapPoll(id);
    }
  });
}

function buildTile(p) {
  const f = printerFields(p);
  const st = f.streamZustand, isLive = st==='live', isSnap = st==='snap';

  // The image is the only real difference between tile and row.
  // So it is here as an add-on and not in the shared fields.
  let cam='';
  const camAttrs=' ondblclick="openVideoFullscreen(\''+esc(p.ip)+'\')" onclick="event.stopPropagation();tileCamClick(\''+esc(p.ip)+'\')" class="clickable"';
  const g2startphase = !G2online && (G2_RESTARTING || (Date.now()-APP_START < 45000));
  const inRepair = isRepair(p.ip);
  const offline = !f.online;
  // Merken, welchen Kamera-Zustand diese Kachel zeigt — daran erkennt
  // refreshTile() later whether a rebuild is needed (spinner out/in).
  const kamAus = isCameraOff(p.ip);
  CAM_KEY[p.ip] = inRepair ? 'repair' : (kamAus ? 'private' : (offline ? 'offline' : 'online'));
  // Order matters: "in repair" and "offline" do NOT depend on the
  // camera server. The printer itself knows whether it is reachable — so
  // show text immediately, regardless of whether go2rtc is still starting. Otherwise the
  // "go2rtc starting" spinner also spun over long-offline printers.
  if(inRepair) cam='<div class="cam-repair"'+camAttrs+'><div class="cip">'+t('inRepair')+'</div></div>';
  else if(kamAus) cam='<div class="cam-repair cam-private"'+camAttrs+'><div class="cip">'+t('camPrivate')+'</div></div>';
  else if(offline) cam='<div class="cam-repair cam-offline"'+camAttrs+'><div class="cip">'+t('lblOffline')+'</div></div>';
  else if(!G2online && g2startphase) cam='<div class="cam-wait"'+camAttrs+'><div class="cam-spinner"></div><div class="cip">'+t('camStarting')+'</div></div>';
  else if(!G2online) cam='<div class="cam-idle"'+camAttrs+'><div class="ci">📷</div><div class="cip">'+esc(p.ip)+'<br>'+t('tileSeeSetup')+'</div></div>';
  else if(isLive) cam='<video class="tile-video" id="vid-'+esc(p.ip)+'" autoplay muted playsinline'+camAttrs+'></video>'
    +'<div class="cam-connect" id="conn-'+esc(p.ip)+'"><div class="cam-spinner"></div><div class="cip">'+t('camConnecting')+'</div></div>';
  else if(isSnap) cam='<div class="snap-container"'+camAttrs+'><img id="snapimg-'+esc(p.ip)+'" src="" class="fill-media" onerror="this.classList.add(\'hidden\');">'
    +'<div class="snap-ring" id="snapring-'+esc(p.ip)+'" title="lädt nächstes Bild"></div>'
    +'<div class="snap-stamp" id="snapts-'+esc(p.ip)+'"></div></div>';
  else cam='<div class="cam-idle ready"'+camAttrs+'><div class="ci">▶</div><div class="cip">'+t('camReady')+'</div></div>';

  return '<div class="df-item kachel tile '+(f.fav?'fav ':'')+(isLive?'streaming ':'')
      +(f.hatFehler?'hat-fehler has-error ':'')+(f.pausiert?'pausiert ':'')
      +(currentViewProfile==='minimal'?'vp-minimal':'')+'" data-ip="'+esc(p.ip)+'">'
    +'<div class="df-kopf">'
      +'<div class="tile-name-wrap"><div class="tile-name clickable" onclick="event.stopPropagation();toggleDD(\''+esc(p.ip)+'\')">'
        +feldHTML(f,'name')+'<span class="tile-name-arrow" id="arr-'+esc(p.ip)+'"> mehr</span></div></div>'
      +feldHTML(f,'stern')
    +'</div>'
    +'<div class="df-pillen">'+feldHTML(f,'online')+feldHTML(f,'status')
      +(isLive?'<span class="tile-live-badge'+(!G2online?' dim':'')+'">LIVE</span>':'')
      +feldHTML(f,'modell')+'</div>'
    +'<div class="tile-cam">'+cam+((isLive||isSnap)?'<button class="tile-fs" title="'+esc(t('videoFullscreen'))+'" onclick="event.stopPropagation();videoFullscreen(\''+esc(p.ip)+'\')">⛶</button>':'')+'</div>'
    +'<div class="df-balken breit"><span class="df-balken-fuell '+esc(f.hatFehler?'failed':f.zustand)
      +'" style="--w:'+f.prozent+'%"></span></div>'
    +'<div class="df-fuss">'
      +'<div class="df-fuss-r1">'+feldHTML(f,'datei')+'</div>'
      +'<div class="df-fuss-r2">'+feldHTML(f,'restzeit')+feldHTML(f,'fortschritt')+'</div>'
    +'</div>'
    +'</div>';
}


// ─── EINE DATENQUELLE JE DRUCKER ──────────────────────────────────────────────
//
// Tile and row showed the same values but fetched them separately:
// buildTile read STATUS directly, renderList again, updateStatusBar and
// updateListRow a third and fourth time. Every change had to be updated in four
// places — and exactly there the bugs of the
// last days arose (the LIVE badge pointing at a removed field; the
// filter that never reached the list).
//
// printerFields is from now on the only place where raw data becomes
// display values. Whoever wants to show something asks here.
// printFileName returns the print file name. Source is subtask_name; if it is
// empty (some models, e.g. woobly 16), the base name from gcode_file is
// used. The extensions .gcode/.3mf are stripped.
function printFileName(s){
  let n = String((s && s.subtask_name) || '').trim();
  if(!n){
    const gf = String((s && s.gcode_file) || '').trim();
    if(gf){ n = gf.split(/[\/\\]/).pop(); }
  }
  return n.replace(/\.gcode$/i,'').replace(/\.3mf$/i,'');
}
// ─── FEHLER QUITTIEREN (roter Rand) ───────────────────────────────────────────
// A printer keeps its red border while it has an error (HMS, print_error or a
// FAILED state). The user can acknowledge it so the border goes away. The
// acknowledgement stores a SIGNATURE of the current error state — as soon as the
// state changes (new error, a new print, a different code) the signature no
// longer matches and the border returns on its own.
let ERR_ACK = {};
try { ERR_ACK = JSON.parse(localStorage.getItem('df_err_ack') || '{}') || {}; } catch(e) { ERR_ACK = {}; }
function rawHatFehler(ip){
  const s = STATUS[ip] || {};
  const errs = ERRORS ? ERRORS.find(e => e.ip === ip) : null;
  const state = String(s.gcode_state || '').toUpperCase();
  return !!(errs && errs.errors && errs.errors.length) || (s.print_error > 0) || state === 'FAILED';
}
function errSig(ip){
  const s = STATUS[ip] || {};
  const errs = ERRORS ? ERRORS.find(e => e.ip === ip) : null;
  const codes = (errs && errs.errors) ? errs.errors.map(e => e.code || e.desc || '').join(',') : '';
  return [String(s.gcode_state || '').toUpperCase(), s.print_error || 0, codes].join('|');
}
function isErrAcked(ip){ return rawHatFehler(ip) && ERR_ACK[ip] === errSig(ip); }
function ackError(ip){
  ERR_ACK[ip] = errSig(ip);
  try { localStorage.setItem('df_err_ack', JSON.stringify(ERR_ACK)); } catch(e){}
  refreshTile(ip);
  // The grid tile's state pill (FAILED) is only rebuilt on a full tile redraw —
  // force it so the tag disappears immediately, not just the border.
  if (typeof renderTile === 'function' && document.querySelector('.tile[data-ip="' + CSS.escape(ip) + '"]')) renderTile(ip);
  toast(t('mdAckDone'), 'ok');
}

function printerFields(p) {
  const s = STATUS[p.ip] || {};
  const errs = ERRORS ? ERRORS.find(e => e.ip === p.ip) : null;
  const rawState = String(s.gcode_state || '').toUpperCase();
  const acked = isErrAcked(p.ip);
  // Once acknowledged, a FAILED state no longer shows its tag/pill — the user
  // has taken note. Everything else follows the real state.
  const state = (acked && rawState === 'FAILED') ? '' : rawState;
  const zust = state.toLowerCase();
  const fehlerText = (errs && errs.errors && errs.errors[0] && errs.errors[0].desc) || '';
  const hatFehler = rawHatFehler(p.ip) && !acked;
  const online = !!s.online;
  const druckt = online && !!zust && zust !== 'idle';

  return {
    ip: p.ip,
    name: p.name || p.ip,
    modell: p.model || '',
    mc: modelColor(p.model),
    fav: !!p.fav,
    serial: p.serial || '',
    ohneSerial: !p.serial,

    online: online,
    zustand: zust,                       // running | pause | failed | finish | idle | ''
    zustandKurz: {running:'PRINTING', pause:'PAUSED', failed:'FAILED', finish:'DONE'}[zust]
                 || (zust ? zust.toUpperCase() : ''),
    druckt: druckt,
    pausiert: zust === 'pause',
    fertig: zust === 'finish',

    datei: printFileName(s),
    prozent: druckt ? (s.progress || 0) : 0,
    prozentBekannt: typeof s.progress === 'number',
    restSek: s.remain_time > 0 ? s.remain_time * 60 : 0,
    restText: s.remain_time > 0 ? fmtRemaining(s.remain_time * 60) : '',

    hatFehler: hatFehler,
    fehlerText: fehlerText,
    streamZustand: ST[p.ip] || 'idle',
  };
}

// feldHTML creates a single display field. Tile and row call the same
// function — so both views are guaranteed the same value, and
// a change takes effect in both at once. The layout is then done by the
// stylesheet, not this code.
// natCmp sorts naturally: "Drucker 2" comes before "Drucker 10". Without it
// 10 would sit right after 1, because comparison would be purely alphabetical.
function natCmp(a, b) {
  return String(a == null ? '' : a).localeCompare(String(b == null ? '' : b),
    undefined, {numeric: true, sensitivity: 'base'});
}

// onlinePill shows after the name whether the printer is currently reachable.
function onlinePill(ip) {
  const s = STATUS[ip];
  const on = !!(s && s.online);
  return '<span class="df-online ' + (on ? 'on' : 'off') + '">' + t(on ? 'lblOnline' : 'lblOffline') + '</span>';
}

// repairPill shows a badge when the printer is marked "in repair".
function repairPill(ip) {
  const s = STATUS[ip];
  if (!(s && s.reparatur && s.reparatur.in_repair)) return '';
  const info = ((s.reparatur.by || '') + ' ' + (s.reparatur.note || '')).trim();
  return '<span class="df-repair" title="' + esc(info) + '">🔧 ' + t('inRepair') + '</span>';
}

// isRepair: true when the printer is currently marked "in repair".
function isRepair(ip) {
  const s = STATUS[ip];
  return !!(s && s.reparatur && s.reparatur.in_repair);
}

// buildRepairBtn: toggle in the detail menu.
function buildRepairBtn(p) {
  const s = STATUS[p.ip];
  const inR = !!(s && s.reparatur && s.reparatur.in_repair);
  return '<div class="dd-sep"></div>'
    + '<button class="hbtn btn-sm ' + (inR ? 'green' : 'red') + '" onclick="toggleRepair(\'' + esc(p.ip) + '\',' + (!inR) + ')">'
    + (inR ? t('repairEnd') : t('repairSet')) + '</button>';
}

// qlRepairBtn: small toggle in the "Printers" list (Saved printers).
function qlRepairBtn(p) {
  const s = STATUS[p.ip];
  const inR = !!(s && s.reparatur && s.reparatur.in_repair);
  return '<button class="hbtn btn-sm ' + (inR ? 'green' : '') + '" onclick="toggleRepair(\'' + esc(p.ip) + '\',' + (!inR) + ')">' + (inR ? '✓ ' + t('qlRepair') : t('qlRepair')) + '</button>';
}

function qlCameraBtn(p){
  const aus = isCameraOff(p.ip);
  return '<button class="hbtn btn-sm'+(aus?' green':'')+'" onclick="qlToggleCamera(\''+esc(p.ip)+'\')">'+(aus?t('qlCamOn'):t('qlCamOff'))+'</button>';
}
function qlBlinkBtn(p){
  if(!isBlinkModel(p.model)) return '';
  const aus = isBlinkOff(p.ip);
  return '<button class="hbtn btn-sm'+(aus?' green':'')+'" onclick="qlToggleBlink(\''+esc(p.ip)+'\')">'+(aus?t('qlBlinkOn'):t('qlBlinkOff'))+'</button>';
}
async function qlToggleCamera(ip){ try{ await toggleCamera(ip);}catch(e){} renderQL(); }
async function qlToggleBlink(ip){ try{ await toggleBlink(ip);}catch(e){} renderQL(); }

async function toggleRepair(ip, an) {
  let note = '';
  if (an) { note = prompt(t('repairNote')) || ''; }
  try {
    const r = await fetch('/api/repair', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ip, in_repair: an, note})});
    const d = await r.json();
    if (STATUS[ip]) STATUS[ip].reparatur = an ? {in_repair:true, note, by:'', since:''} : {in_repair:false};
    refreshTile(ip);
    renderQL();
    if (DD_OPEN_IP === ip) {
      const pop = document.getElementById('dd-pop'); const p = P.find(x => x.ip === ip);
      if (pop && p) pop.innerHTML = '<div class="dd-title">' + esc(p.name) + '<span class="dd-close" title="' + esc(t('close')) + '" onclick="event.stopPropagation();closeDD()">✕</span></div>' + buildDropdownContent(p);
    }
    toast(an ? t('repairSetOk') : t('repairEndOk'), 'ok');
    if (an && d && d.auf_sd === false) toast(t('repairOffline'), 'er');
  } catch(e) { toast('Fehlgeschlagen', 'er'); }
}

function feldHTML(f, welches) {
  switch (welches) {
    case 'stern':
      // Same glyph width in both states: ★ and ☆ come from
      // the same font family, whereas ⭐ is an emoji and comes out larger.
      return '<span class="df-stern' + (f.fav ? ' on' : '') + '" data-ip="' + esc(f.ip)
           + '" onclick="event.stopPropagation();tileToggleFav(this.dataset.ip)">'
           + (f.fav ? '★' : '☆') + '</span>';

    case 'modell':
      return '<span class="df-modell" style="--mc:' + f.mc + '">' + esc(f.modell) + '</span>';

    case 'name':
      // The pill MUST sit in the same .df-name wrapper: refreshTile() replaces
      // only .df-name on a status tick — if the pill were beside it, a new one would
      // be added on every tick.
      return '<span class="df-name"><span class="df-name-txt">' + esc(f.name) + '</span>' + fwPillHTML(f.ip) + repairPill(f.ip) + '</span>';

    case 'online':
      // The pill is clickable: a click pings the printer (ports 322/990/8883)
      // — the same check as in the "+ Printers" list, but right here.
      return '<span class="df-online-wrap ql-ping" title="' + esc(t('pingTitle')) + '" onclick="event.stopPropagation();pingPrinter(\'' + esc(f.ip) + '\')">' + onlinePill(f.ip) + '</span>';

    case 'datei':
      if (f.hatFehler && f.fehlerText)
        return '<span class="df-datei fehler">⚠ ' + esc(f.fehlerText) + '</span>';
      return '<span class="df-datei">' + (f.datei ? esc(f.datei) : '–') + '</span>';

    case 'fortschritt':
      if (f.hatFehler)
        return '<span class="df-fortschritt">'
             + '<span class="df-balken"><span class="df-balken-fuell failed" style="--w:' + f.prozent + '%"></span></span>'
             + '<span class="df-prozent">✗ ' + f.prozent + '%</span></span>';
      if (!f.druckt && !f.prozent)
        return '<span class="df-fortschritt"><span class="df-leer">'
             + esc(f.zustandKurz || (f.online ? 'bereit' : 'offline')) + '</span></span>';
      return '<span class="df-fortschritt">'
           + '<span class="df-balken"><span class="df-balken-fuell ' + esc(f.zustand)
           + '" style="--w:' + f.prozent + '%"></span></span>'
           + '<span class="df-prozent">' + f.prozent + '%</span></span>';

    case 'restzeit':
      return '<span class="df-rest">' + esc(f.restText) + '</span>';

    case 'ip':
      return '<span class="df-ip">' + esc(f.ip) + '</span>';

    case 'status':
      if (!f.zustandKurz || !f.online) return '<span class="df-status"></span>';
      return '<span class="df-status ' + esc(f.zustand) + '">' + esc(f.zustandKurz) + '</span>';
  }
  return '';
}

// The field order is the same for both views. To add a column,
// add it here — once.
const DF_FELDER = ['stern', 'modell', 'name', 'datei', 'fortschritt', 'restzeit', 'ip'];

function felderHTML(f) {
  return DF_FELDER.map(n => feldHTML(f, n)).join('');
}

// ─── VIEW PROFILE ─────────────────────────────────────────────────────────────
let currentViewProfile = localStorage.getItem('viewProfile') || 'regular';

function toggleViewProfile(e) {
  e.stopPropagation();
  const menu = document.getElementById('viewprofile-menu');
  if (!menu) return;
  const btn = document.getElementById('tab-viewprofile');
  const r = btn.getBoundingClientRect();
  menu.style.top = r.bottom + 'px';
  menu.style.left = r.left + 'px';
  setShown(menu, menu.classList.contains('hidden'));
}
document.addEventListener('click', () => {
  const m = document.getElementById('viewprofile-menu');
  if (m) setShown(m, false);
});

function setViewProfile(profile) {
  currentViewProfile = profile;
  localStorage.setItem('viewProfile', profile);
  const vpr = byId('vp-regular'), vpm = byId('vp-minimal');
  if (vpr) vpr.classList.toggle('active', profile === 'regular');
  if (vpm) vpm.classList.toggle('active', profile === 'minimal');
  setShown(document.getElementById('viewprofile-menu'), false);
  // Apply to all tiles
  document.querySelectorAll('.tile').forEach(t => {
    t.classList.toggle('vp-minimal', profile === 'minimal');
  });
  // Re-render to apply
  renderGrid();
}


// ─── FULLSCREEN VIDEO ONLY ────────────────────────────────────────────────────

// The tile IDs (vid-<ip>, snapimg-<ip>) are globally unique. Before a
// fullscreen view creates the same IDs, the grid must be cleared — otherwise
// getElementById returns the invisible element from the grid and the
// fullscreen view builds a second stream. Exactly that was the reason
// for the double load when switching to fullscreen.
function clearGridDOM() {
  const g = document.getElementById('printerGrid');
  if (g) g.innerHTML = '';
}

// fsCell renders a bare fullscreen cell in the currently selected display mode.
// fsHinweisHTML: in fullscreen show text instead of spinner/image when the
// printer is offline or in repair — just like in the overview.
function fsHinweisHTML(ip) {
  if (isRepair(ip)) return '<div class="fsv-hinweis repair"><span>' + t('inRepair') + '</span></div>';
  if (isCameraOff(ip)) return '<div class="fsv-hinweis private"><span>' + t('camPrivate') + '</span></div>';
  const s = STATUS[ip];
  if (!(s && s.online)) return '<div class="fsv-hinweis offline"><span>' + t('lblOffline') + '</span></div>';
  return '';
}

function fsCell(p, cellStyle) {
  const hinweis = fsHinweisHTML(p.ip);
  if (hinweis) return '<div class="'+cellStyle+'" data-ip="'+esc(p.ip)+'">' + hinweis + '</div>';
  const inner = SNAP_MODE > 0
    ? '<img id="snapimg-'+esc(p.ip)+'" src="" class="fill-media" onload="hideFsSpin(\''+esc(p.ip)+'\')" onerror="this.classList.add(\'hidden\');">'
      + '<div class="snap-stamp" id="snapts-'+esc(p.ip)+'"></div>'
    : '<video id="vid-'+esc(p.ip)+'" autoplay muted playsinline class="fill-media"></video>';
  return '<div class="'+cellStyle+'" data-ip="'+esc(p.ip)+'">' + inner
    + '<div class="fs-spin" id="fsspin-'+esc(p.ip)+'"><div class="cam-spinner"></div></div>'
    + '<button class="tile-fs fsv-cell-fs" title="'+esc(t('videoFullscreen'))+'" onclick="event.stopPropagation();cellFullscreen(\''+esc(p.ip)+'\')">\u26db</button>'
    + '</div>';
}

// fsvCell is the cell for "Fullscreen Video Only": image only, a frame in
// status color and a progress bar at the bottom — deliberately NO text.
// hideFsSpin hides a fullscreen cell's spinner once an image is present.
function hideFsSpin(ip) {
  const el = document.getElementById('fsspin-' + ip);
  if (el) el.remove();
}

function fsvCell(p) {
  const f = printerFields(p);
  const zust = f.hatFehler ? 'failed' : (f.zustand || 'idle');
  const w = f.prozent || 0; // Balken bleibt bei Fehler/Pause sichtbar
  const hinweis = fsHinweisHTML(p.ip);
  const inner = hinweis ? hinweis : (
      (SNAP_MODE > 0
        ? '<img id="snapimg-'+esc(p.ip)+'" src="" class="fill-media" onload="hideFsSpin(\''+esc(p.ip)+'\')" onerror="this.classList.add(\'hidden\');">'
        : '<video id="vid-'+esc(p.ip)+'" autoplay muted playsinline class="fill-media"></video>')
      + '<div class="fs-spin" id="fsspin-'+esc(p.ip)+'"><div class="cam-spinner"></div></div>');
  return '<div class="scan-cell fsv-cell '+esc(zust)+'" data-ip="'+esc(p.ip)+'">'
    + '<div class="fsv-head"><div class="fsv-name">'+esc(p.name)+'</div>'
    + (f.restText ? '<div class="fsv-sub">'+esc(f.restText)+'</div>' : '') + '</div>'
    + inner
    + '<div class="fsv-bar"><span class="fsv-bar-fill '+esc(zust)+'" style="--w:'+w+'%"></span></div>'
    + '</div>';
}

// updateFSVCell keeps a cell's frame color and bar current (live status).
function updateFSVCell(ip, f) {
  const soll = isRepair(ip) ? 'repair' : (isCameraOff(ip) ? 'private' : (!f.online ? 'offline' : 'online'));
  const kat = el => el.querySelector('.fsv-hinweis.repair') ? 'repair'
                  : el.querySelector('.fsv-hinweis.private') ? 'private'
                  : el.querySelector('.fsv-hinweis.offline') ? 'offline' : 'online';
  const p = P.find(x => x.ip === ip);

  // „Video Only" (view-fsv)
  const c = document.querySelector('.fsv-cell[data-ip="'+CSS.escape(ip)+'"]');
  if (c) {
    if (soll !== kat(c) && p) {
      if (soll !== 'online') { stopWebRTC(ip); stopSnapPoll(ip); }
      c.outerHTML = fsvCell(p);
      if (soll === 'online' && (ST[ip] === 'live' || ST[ip] === 'snap')) mountStreams([p]);
    } else {
      const zust = f.hatFehler ? 'failed' : (f.zustand || 'idle');
      c.className = 'scan-cell fsv-cell ' + zust;
      const fill = c.querySelector('.fsv-bar-fill');
      if (fill) { fill.className = 'fsv-bar-fill ' + zust; fill.style.setProperty('--w', (f.prozent||0)+'%'); }
    }
  }

  // Das separate Vollbild-Overlay (fsCell / scan-cell-fixed)
  const fixed = document.querySelector('.scan-cell-fixed[data-ip="'+CSS.escape(ip)+'"]');
  if (fixed && soll !== kat(fixed) && p) {
    if (soll !== 'online') { stopWebRTC(ip); stopSnapPoll(ip); }
    fixed.outerHTML = fsCell(p, 'scan-cell-fixed');
    if (soll === 'online' && (ST[ip] === 'live' || ST[ip] === 'snap')) mountStreams([p]);
  }
}

// adoptForFullscreen makes the fullscreen view adopt the same mode
// active in the grid, instead of opening its own.
function adoptForFullscreen(printers) {
  printers.forEach(p => { if ((ST[p.ip] || 'idle') === 'idle') setTileState(p.ip, activeMode()); });
}

function startFullscreenVideoOnly() {
  let overlay = document.getElementById('fsvonly-overlay');
  if (!overlay) {
    overlay = document.createElement('div');
    overlay.id = 'fsvonly-overlay';
    overlay.className = 'fsv-overlay hidden';
    overlay.title = 'Klicken oder ESC zum Schliessen';
    overlay.addEventListener('click', closeFullscreenVideoOnly);
    document.body.appendChild(overlay);
  }
  const cols = parseInt(document.getElementById('colR').value) || 9;
  const printers = getFlt();

  clearGridDOM();
  adoptForFullscreen(printers);

  overlay.innerHTML = printers.map(p =>
    fsCell(p, 'scan-cell-fixed')
  ).join('');
  setShown(overlay, true);
  overlay.style.setProperty('--cols', cols);

  requestAnimationFrame(() => mountStreams(printers));
  document.addEventListener('keydown', fsvOnlyEsc);
}

function closeFullscreenVideoOnly() {
  const overlay = document.getElementById('fsvonly-overlay');
  if (!overlay) return;
  setShown(overlay, false);
  overlay.innerHTML = ''; // IDs freigeben, bevor das Grid sie wieder vergibt
  document.removeEventListener('keydown', fsvOnlyEsc);
  resumeAllTiles();
}

function fsvOnlyEsc(e) {
  if (e.key === 'Escape') closeFullscreenVideoOnly();
}

// ─── SENDESEITE (Broadcast) ───────────────────────────────────────────────────
// A separate, slim start for /broadcast: only load data and the grid
// zeichnen — keine Bedienoberflaeche, keine Steuer-Buttons.
let BC_LAST = '';

function broadcastList() {
  const opts = window.BROADCAST_OPTS || {};
  const view = opts.view || 'all';
  let list = P.slice();
  if (view === 'online') list = list.filter(p => STATUS[p.ip] && STATUS[p.ip].online);
  else if (view === 'offline') list = list.filter(p => !(STATUS[p.ip] && STATUS[p.ip].online));
  return (typeof sortList === 'function') ? sortList(list) : list;
}

function renderBroadcast() {
  const opts = window.BROADCAST_OPTS || {};
  const list = broadcastList();
  const cont = document.getElementById('bc-grid');
  if (!cont) return;
  list.forEach(p => { if ((ST[p.ip] || 'idle') === 'idle') setTileState(p.ip, activeMode()); });
  cont.style.setProperty('--cols', opts.cols || 9);
  cont.innerHTML = list.map(p => fsvCell(p)).join('');
  BC_LAST = list.map(p => p.ip).join(',');
  requestAnimationFrame(() => mountStreams(list));
}

async function bootBroadcast() {
  const opts = window.BROADCAST_OPTS || {};
  CURRENT_SUBVIEW = 'broadcast';
  document.body.classList.add('broadcast-mode');
  if (opts.click === false) document.body.classList.add('bc-noclick');
  try { await loadSettings(); } catch(e) {}
  applyLang();
  try { await loadPrinters(); } catch(e) {}
  try { await loadCameras(); } catch(e) {}
  try { await loadStatus(); } catch(e) {}
  if (opts.remote) {
    // Over the tunnel only snapshot works (live WebRTC does not traverse it).
    SNAP_MODE = 2; SNAP_AUTO = true;
  } else {
    try { await checkComponentsOnStart(); } catch(e) {}
    const have = k => COMPONENTS.some(c => c.key === k && c.installed);
    if (have('go2rtc') && have('ffmpeg')) { SNAP_MODE = 2; SNAP_AUTO = true; }
  }
  renderBroadcast();
  tickers.push(setInterval(async () => { try { await loadStatus(); } catch(e) {} }, 30000));
  const sp = document.getElementById('splash'); if (sp) sp.style.display = 'none';
}

function tileCamClick(ip) {
  if (SNAP_MODE > 0) {
    if (ST[ip] === 'snap') {
      setTileState(ip, 'idle');
      renderTile(ip);
    } else {
      setTileState(ip, 'snap');
      renderTile(ip); // das <img> muss existieren, bevor die erste Antwort kommt
      startSnapPoll(ip);
    }
  } else {
    if (ST[ip] === 'live') { tileStop(ip); } else { tilePlay(ip); }
  }
}

// videoFullscreen switches EXACTLY this video/image of a tile into the fullscreen
// modus des Browsers (Knopf unten rechts im Video).
function videoFullscreen(ip) {
  const cam = document.querySelector('.tile[data-ip="' + CSS.escape(ip) + '"] .tile-cam');
  if (!cam) return;
  const el = cam.querySelector('video') || cam.querySelector('img.fill-media') || cam.querySelector('.snap-container') || cam;
  if (el.requestFullscreen) el.requestFullscreen().catch(()=>{});
  else if (el.webkitRequestFullscreen) el.webkitRequestFullscreen();
}

// cellFullscreen switches ONE cell from the overall fullscreen overlay into the
// real browser fullscreen mode. stopPropagation in the button prevents the
// click from closing the overlay.
function cellFullscreen(ip) {
  const cell = document.querySelector('#fsvonly-overlay .scan-cell-fixed[data-ip="' + CSS.escape(ip) + '"]')
            || document.querySelector('.fsv-cell[data-ip="' + CSS.escape(ip) + '"]')
            || document.querySelector('.scan-cell-fixed[data-ip="' + CSS.escape(ip) + '"]');
  if (!cell) return;
  const el = cell.querySelector('video') || cell.querySelector('img.fill-media') || cell;
  if (el.requestFullscreen) el.requestFullscreen().catch(()=>{});
  else if (el.webkitRequestFullscreen) el.webkitRequestFullscreen();
}

function openVideoFullscreen(ip) {
  const cam = document.getElementById('tc-'+ip);
  if (!cam) return;
  const video = cam.querySelector('video');
  const el = video || cam;
  if (el.requestFullscreen) el.requestFullscreen();
  else if (el.webkitRequestFullscreen) el.webkitRequestFullscreen();
}
document.addEventListener('fullscreenchange', () => {
  if (!document.fullscreenElement) {
    // Exited fullscreen
  }
});



function renderGrid(){
  const el=document.getElementById('printerGrid'), F=getFlt();
  const sb=document.getElementById('grid-sortbar'); if(sb) sb.innerHTML=gridSortHTML();
  if(!P.length && !CAMS.length){el.innerHTML='<div class="skeleton"><div class="spinner"></div><div>'+t('noDevices')+'</div><div>'+t('noDevicesHint')+'</div></div>';return;}
  if(!F.length && !CAMS.length){el.innerHTML='<div class="skeleton"><div>'+t('noResults')+'</div><div>'+t('noResultsHint')+'</div></div>';return;}
  if(SNAP_MODE>0 && SNAP_AUTO) F.forEach(p=>{ if((ST[p.ip]||'idle')==='idle') setTileState(p.ip,'snap'); });
  // Cameras get the same start mode as the tiles (standalone, appended at the end).
  CAMS.forEach(c=>{ if((ST[c.id]||'idle')==='idle') setTileState(c.id, activeMode()); });
  el.innerHTML=F.map(p=>buildTile(p)).join('') + CAMS.map(c=>buildCamTile(c)).join('');
  mountStreams(F);
  mountCamStreams();
}
function renderTile(ip){
  const ex=document.querySelector('.tile[data-ip="'+ip+'"]');
  if(!ex)return;
  const p=P.find(x=>x.ip===ip);
  if(p){ ex.outerHTML=buildTile(p); return; }
  const cam=CAMS.find(c=>c.id===ip);
  if(cam){ ex.outerHTML=buildCamTile(cam); }
}

// LIST_AUF remembers which detail rows are open. Without this memory
// everything collapsed on each status message — the list is redrawn every few
// seconds.
let LIST_AUF = new Set();

function listToggle(ip, ev) {
  if (ev) {
    // Clicks on the star or on buttons in the detail row not counted as
    // open-and-close.
    const t = ev.target;
    if (t.closest && (t.closest('.lstar') || t.closest('.list-detail') || t.closest('button'))) return;
  }
  if (LIST_AUF.has(ip)) LIST_AUF.delete(ip); else LIST_AUF.add(ip);
  drawListRow(ip);
}

// listAlleUm decides on click itself which direction it goes.
// Previously the direction was fixed in the onclick attribute, which on toggle
// was not updated — after expanding once the button forever called
// "expand" and "collapse all" did nothing.
function listAlleUm() {
  const sichtbar = getFlt();
  const alleOffen = sichtbar.length > 0 && sichtbar.every(p => LIST_AUF.has(p.ip));
  listAlle(!alleOffen);
}

function listAlle(auf) {
  const sichtbar = getFlt();
  LIST_AUF = auf ? new Set(sichtbar.map(p => p.ip)) : new Set();
  sichtbar.forEach(p => drawListRow(p.ip));
  const b = byId('btn-list-alle');
  if (b) b.textContent = auf ? '▾ Alle einklappen' : '▸ Alle ausklappen';
}

// drawListRow toggles a single row without rebuilding the whole list
// — otherwise the whole table flickers on expand.
// detailFrei says whether the detail row may be safely rebuilt.
// Three reasons against: a form is open, the cursor is in an
// input field, or a select list is expanded.
function detailFrei(ip) {
  const det = byId('ld-' + ip);
  if (!det) return false;
  if (det.dataset.halten === '1') return false;
  const a = document.activeElement;
  if (a && a !== document.body && det.contains(a)) return false;
  return true;
}

function detailHalten(ip, an) {
  const det = byId('ld-' + ip);
  if (det) det.dataset.halten = an ? '1' : '0';
}

function drawListRow(ip) {
  const zeile = byId('li-' + ip), det = byId('ld-' + ip);
  const auf = LIST_AUF.has(ip);
  if (zeile) zeile.classList.toggle('auf', auf);
  if (!det) return;
  det.classList.toggle('auf', auf);

  if (!auf) {
    // Collapsed means: fetch no more images. Otherwise the list kept running
    // unnoticed queries for tiles nobody sees.
    det.innerHTML = '';
    listImageOff(ip);
    return;
  }

  const p = P.find(x => x.ip === ip);
  if (!p) return;
  det.innerHTML = detailKaesten(p);
  listImageOn(ip);
}

// ─── KAMERA-ADRESSE MESSEN ────────────────────────────────────────────────────
//
// Which address form a device understands is decided by its firmware. Instead of
// deriving it from the model name, this button tries the forms on the device
// and remembers the one that returns an image.
async function cameraTest(ip) {
  const feld = byId('kt-' + ip);
  if (feld) feld.innerHTML = '<span class="spinner spinner-sm"></span> probiert Adressen …';
  busyStart('testet Kamera …');
  try {
    const d = await (await fetch('/api/camera-test?ip=' + encodeURIComponent(ip), {cache:'no-store'})).json();
    if (feld) {
      const liste = (d.ergebnisse || [])
        .map(e => (e.ok ? '✓ ' : '✗ ') + e.schema + e.pfad + (e.grund ? ' (' + e.grund + ')' : ''))
        .join(' · ');
      feld.innerHTML = '<span class="' + (d.gemerkt ? 'kt-gut' : 'kt-schlecht') + '">'
        + esc(d.urteil || '') + '</span>' + (liste ? '<br><span class="dd-fein">' + esc(liste) + '</span>' : '');
    }
    if (d.gemerkt) toast('Adresse gemerkt — go2rtc wurde neu eingerichtet', 'ok');
  } catch(e) {
    if (feld) feld.textContent = e.message || String(e);
  } finally { busyEnd(); }
}

// ─── DIE VIER KAESTEN DER DETAILANSICHT ───────────────────────────────────────
//
// Video, print, printer, AMS — each group on its own, so that at a glance one
// knows where to look. Previously everything was in one long list.

function kasten(titel, inhalt, extraKlasse) {
  if (!inhalt) return '';
  return '<div class="dk ' + (extraKlasse || '') + '">'
       + '<div class="dk-titel">' + esc(titel) + '</div>'
       + '<div class="dk-inhalt">' + inhalt + '</div></div>';
}

function dkRow(bezeichnung, wert, klasse) {
  if (wert === '' || wert === undefined || wert === null) return '';
  return '<div class="dk-zeile"><span class="dk-bez">' + esc(bezeichnung) + '</span>'
       + '<span class="dk-wert ' + (klasse || '') + '">' + wert + '</span></div>';
}

function detailKaesten(p) {
  const f = printerFields(p);
  const st = STATUS[p.ip] || {};

  // ── Video ──
  const video = '<div class="ld-cam" id="ldcam-' + esc(p.ip) + '">'
    + '<img id="ldsnap-' + esc(p.ip) + '" class="fill-media" src="" onerror="this.classList.add(\'hidden\');">'
    + '<div class="ld-cam-warten" id="ldwarten-' + esc(p.ip) + '"><span class="spinner"></span></div>'
    + '</div>';

  // ── Druck ──
  let druck = '';
  if (st.online) {
    druck += dkRow('Datei', f.datei ? esc(f.datei) : '–');
    druck += dkRow('Fortschritt', f.druckt || f.prozent
      ? '<span class="dk-balken"><span class="dk-balken-fuell ' + esc(f.zustand)
        + '" style="--w:' + f.prozent + '%"></span></span> ' + f.prozent + '%'
      : esc(f.zustandKurz || 'bereit'));
    druck += dkRow('Verbleibend', f.restText ? esc(f.restText) : '–');
    druck += dkRow('Schicht', (st.layer > 0 || st.total_layers > 0)
      ? esc(String(st.layer || 0)) + ' / ' + esc(String(st.total_layers || 0)) : '–');
    druck += buildPrintControls(p);
  } else {
    druck = '<div class="dk-leer">' + (p.serial ? 'Offline' : 'Keine Seriennummer') + '</div>';
  }

  // ── Printer ──
  const netzName = st.dev_name || '';
  let drucker = dkRow('Name', esc(p.name)
      + (netzName && netzName !== p.name ? ' <span class="dd-fein">im Gerät: ' + esc(netzName) + '</span>' : ''));
  drucker += dkRow('Modell', esc(p.model || '–'));
  drucker += dkRow('IP', esc(p.ip), 'mono');
  drucker += dkRow('Seriennummer', esc(p.serial || '–'), 'mono');
  drucker += dkRow('Dienstzeit', st.laufzeit_text
      ? esc(st.laufzeit_text) + ' <span class="dd-fein">gemessen seit Eintrag</span>' : '–');
  drucker += dkRow('Firmware', (st.info && st.info.firmware) ? esc(st.info.firmware) : '–', 'mono');
  drucker += dkRow('Favorit', '<span class="df-stern' + (f.fav ? ' on' : '') + '" data-ip="' + esc(p.ip)
      + '" onclick="event.stopPropagation();tileToggleFav(this.dataset.ip)">' + (f.fav ? '★' : '☆') + '</span>');
  // Temperature control belongs to the device, not the running print.
  drucker += buildTempControls(p);

  // ── AMS, hoechstens vier Einheiten ──
  let ams = '';
  const einheiten = (st.ams || []).slice(0, 4);
  const amsInfo = ((st.info && st.info.ams) || []).slice(0, 4);
  // There can be more reported modules than tray data — e.g. when an
  // AMS is connected but has not read any spools yet. Both count,
  // otherwise an existing unit disappears from the display.
  const anzahl = Math.min(4, Math.max(einheiten.length, amsInfo.length));
  for (let i = 0; i < anzahl; i++) {
    const u = einheiten[i] || {};
    const m = amsInfo[i] || {};
    ams += '<div class="dk-ams">'
      + '<div class="dk-ams-kopf">' + esc(amsName(m.name || ('ams/' + i))) + '</div>'
      + dkRow('Modell', esc(m.hw_ver || '–'), 'mono')
      + dkRow('Seriennummer', esc(m.sn || '–'), 'mono')
      + dkRow('Firmware', esc(m.sw_ver || '–'), 'mono')
      // All four trays as chips — the same form as in the grid, so an
      // empty tray is as small as a filled one and not as a column
      // erscheint.
      + '<div class="ams-trays">'
      + [0, 1, 2, 3].map(si => {
          const tr = (u.trays || [])[si] || {};
          const belegt = !!tr.type;
          return '<span class="ams-tray' + (belegt ? '' : ' leer') + '" title="'
            + esc(belegt ? (tr.type + (tr.sub_brand ? ' ' + tr.sub_brand : '')) : 'Fach ' + (si + 1) + ' leer') + '">'
            + amsPunkt(tr) + '<span class="ams-tray-typ">' + esc(belegt ? tr.type : '–') + '</span></span>';
        }).join('')
      + '</div></div>';
  }
  if (!ams && st.ext_spool && st.ext_spool.type) {
    ams = '<div class="dk-zeile"><span class="dk-bez">Externe Rolle</span><span class="dk-wert">'
        + esc(st.ext_spool.type) + '</span></div>';
  }

  const fehler = ERRORS ? ERRORS.find(e => e.ip === p.ip) : null;
  const fehlerBlock = (fehler && fehler.messages && fehler.messages.length)
    ? '<div class="dk-fehler">' + fehler.messages.map(m => esc(m)).join('<br>') + '</div>' : '';

  return '<div class="dk-raster">'
    + kasten('Video', video
        + '<div class="dk-videotest"><button class="hbtn btn-sm" data-ip="' + esc(p.ip)
        + '" onclick="event.stopPropagation();cameraTest(this.dataset.ip)">🔍 Kamera-Adresse testen</button>'
        + '<span class="note-sm" id="kt-' + esc(p.ip) + '"></span></div>', 'dk-video')
    + kasten('Druck', druck + fehlerBlock, 'dk-druck')
    + '<div class="dk dk-drucker"><div class="dk-titel">Drucker</div>'
      + '<div class="dk-inhalt" id="dk-drucker-' + esc(p.ip) + '">' + drucker
      + '<div class="ld-akt"><button class="hbtn btn-sm" data-ip="' + esc(p.ip)
      + '" onclick="event.stopPropagation();bearbeiteInline(this.dataset.ip)">✎ Bearbeiten</button></div>'
      + '</div></div>'
    + kasten('AMS', buildAMSUnits(p) || ams, 'dk-ams')
    + '</div>';
}

// ─── BILD IN DER AUFGEKLAPPTEN ZEILE ──────────────────────────────────────────
//
// Only the expanded row fetches an image — into its own element.
// The tile in the grid does not carry the same id twice; otherwise the
// image query would write into the invisible element in the hidden grid.
let LIST_BILD = {};

function listImageOn(ip) {
  listImageOff(ip);
  const takt = Math.max(2, SNAP_MODE || 5);
  const hol = () => fetchListImage(ip);
  hol();
  LIST_BILD[ip] = setInterval(hol, takt * 1000);
}

function listImageOff(ip) {
  if (LIST_BILD[ip]) { clearInterval(LIST_BILD[ip]); delete LIST_BILD[ip]; }
}

function listImagesAllOff() {
  Object.keys(LIST_BILD).forEach(listImageOff);
}

async function fetchListImage(ip) {
  if (NETZ_PAUSE) return;
  const img = byId('ldsnap-' + ip);
  if (!img) { listImageOff(ip); return; }
  try {
    const r = await fetch('/api/snapshot/' + encodeURIComponent(ip) + '?max_age=' + Math.max(2, SNAP_MODE || 5),
                          {cache: 'no-store'});
    if (!r.ok) {
      const grund = (await r.text()).trim() || ('HTTP ' + r.status);
      const w = byId('ldwarten-' + ip);
      if (w) w.innerHTML = '<span class="ld-cam-fehler">' + esc(grund) + '</span>';
      return;
    }
    const blob = await r.blob();
    if (!blob.size) return;
    const url = URL.createObjectURL(blob);
    const alt = img.dataset.url;
    img.dataset.url = url;
    img.onload = img.onerror = () => { if (alt) URL.revokeObjectURL(alt); };
    img.classList.remove('hidden');
    img.src = url;
    setShown(byId('ldwarten-' + ip), false);
  } catch(e) {}
}

// Editing exists in only one place — the "Printers" tab. Previously there was
// a second version of editPrinter for it, which rebuilt list rows in the old
// layout; it was ineffective since the rework and has been removed.
// Editing happens where the printer is being viewed. The former
// jump to the "Printers" tab was bad for two reasons: one lost the
// spot one was at, and the flow depended on a 60 ms delay
// plus scrollIntoView — both things that can go wrong without it
// jemand merkt.
function editInPrinterTab(ip) { bearbeiteInline(ip); }

function bearbeiteInline(ip) {
  const p = P.find(x => x.ip === ip);
  const kasten = byId('dk-drucker-' + ip);
  if (!p || !kasten) return;
  const feld = (kennung, wert, platzhalter, extra) =>
    '<div class="dk-zeile"><span class="dk-bez">' + esc(platzhalter) + '</span>'
    + '<input class="dk-eingabe ' + (extra || '') + '" id="ei-' + kennung + '-' + esc(ip)
    + '" value="' + esc(wert || '') + '"></div>';
  kasten.innerHTML =
      feld('name', p.name, 'Name')
    + feld('model', p.model, 'Modell')
    + feld('ip', p.ip, 'IP', 'mono')
    + feld('serial', p.serial, 'Seriennummer', 'mono')
    + feld('code', p.code, 'Zugangscode', 'mono')
    + '<div class="ld-akt">'
    + '<button class="hbtn green btn-sm" data-ip="' + esc(ip) + '" onclick="event.stopPropagation();saveInline(this.dataset.ip)">Speichern</button>'
    + '<button class="hbtn btn-sm" data-ip="' + esc(ip) + '" onclick="event.stopPropagation();detailHalten(this.dataset.ip,false);drawListRow(this.dataset.ip)">Abbrechen</button>'
    + '</div>';
  detailHalten(ip, true);
  const n = byId('ei-name-' + ip);
  if (n) { n.focus(); n.select(); }
  kasten.querySelectorAll('input').forEach(i => i.addEventListener('keydown', e => {
    e.stopPropagation();
    if (e.key === 'Enter') saveInline(ip);
    if (e.key === 'Escape') drawListRow(ip);
  }));
}

async function saveInline(ip) {
  const p = P.find(x => x.ip === ip);
  if (!p) return;
  const v = k => ((byId('ei-' + k + '-' + ip) || {}).value || '').trim();
  const name = v('name'), model = v('model').toUpperCase(), code = v('code');
  const serial = v('serial'), neueIP = v('ip') || p.ip;

  if (!name || !model || !code) { toast('Name, Modell und Zugangscode dürfen nicht leer sein', 'er'); return; }
  if (!/^(\d{1,3}\.){3}\d{1,3}$/.test(neueIP) || neueIP.split('.').some(x => parseInt(x, 10) > 255)) {
    toast('IP-Adresse sieht nicht richtig aus', 'er'); return;
  }
  if (neueIP !== p.ip && P.some(x => x.ip === neueIP)) {
    toast('Diese IP gehört schon zu ' + P.find(x => x.ip === neueIP).name, 'er'); return;
  }

  busyStart('speichert …');
  try {
    const r = await fetch('/api/printers/' + encodeURIComponent(ip), {
      method: 'PUT', headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({...p, model, name, ip: neueIP, code, serial})
    });
    if (!r.ok) throw new Error((await r.text()).trim() || ('HTTP ' + r.status));
    const umgezogen = neueIP !== ip;
    if (umgezogen) {
      if (ST[ip]) { setTileState(ip, 'idle'); delete ST[ip]; }
      delete STATUS[ip];
      listImageOff(ip);
      LIST_AUF.delete(ip);
      LIST_AUF.add(neueIP);
    }
    await loadPrinters();
    render();
    detailHalten(ip, false);
    if (CURRENT_SUBVIEW === 'list') renderList();
    toast(umgezogen ? name + ' hört jetzt auf ' + neueIP : name + ' gespeichert', 'ok');
  } catch(e) {
    toast(e.message || String(e), 'er');
  } finally { busyEnd(); }
}

// ─── SORTIERUNG DER LISTE ─────────────────────────────────────────────────────
function sortWert(p, key) {
  const f = printerFields(p);
  const st = STATUS[p.ip] || {};
  switch (key) {
    case 'modell':      return f.modell || '';
    case 'name':        return f.name || '';
    case 'online':      return st.online ? 1 : 0;
    case 'datei':       return f.datei || '';
    case 'fortschritt': return Number(st.remain_time) || 0; // nach Restzeit sortieren
    case 'restzeit':    return Number(st.remain_time) || 0;
    case 'ip':          return p.ip || '';
    default:            return f.name || '';
  }
}

function sortList(list) {
  const key = LIST_SORT.key, dir = LIST_SORT.dir;
  const zahl = (key === 'online' || key === 'fortschritt' || key === 'restzeit');
  return list.slice().sort((a, b) => {
    // When sorting by progress/remaining: printers WITHOUT a running print
    // ("ready"/idle/offline) always at the end — otherwise they sit with remaining 0
    // above the almost-finished prints.
    if (key === 'fortschritt' || key === 'restzeit') {
      const da = printerFields(a).druckt ? 0 : 1;
      const db = printerFields(b).druckt ? 0 : 1;
      if (da !== db) return da - db;
    }
    const va = sortWert(a, key), vb = sortWert(b, key);
    let c = zahl ? (va - vb) : natCmp(va, vb);
    if (c === 0) c = natCmp(a.name || '', b.name || '');  // Zweitschlüssel: Name
    return c * dir;
  });
}

function setListSort(key) {
  if (LIST_SORT.key === key) LIST_SORT.dir = -LIST_SORT.dir;  // gleiche Spalte: Richtung drehen
  else LIST_SORT = {key: key, dir: 1};
  try { localStorage.setItem('listSort', JSON.stringify(LIST_SORT)); } catch(e){} // für den nächsten Start merken
  if (CURRENT_SUBVIEW === 'list') renderList(); else renderGrid();
}

// gridSortHTML builds the sort bar above the grid — the same sorting as
// the list's column header, just as a horizontal bar, because tiles
// have no table headers.
function gridSortHTML() {
  const th = (key, label) => '<span class="sortbar' + (LIST_SORT.key===key?' aktiv':'')
      + '" onclick="setListSort(\''+key+'\')">' + label + listSortPfeil(key) + '</span>';
  return '<span class="gsb-label">Sortieren:</span>'
    + th('name','Name') + th('online','Status') + th('modell','Modell')
    + th('fortschritt','Fortschritt') + th('ip','IP');
}

function listSortPfeil(key) {
  if (LIST_SORT.key !== key) return '';
  return '<span class="pf">' + (LIST_SORT.dir > 0 ? ' ▲' : ' ▼') + '</span>';
}

function renderList(){
  const el=document.getElementById('printerList');
  // The list shows exactly what the filter leaves — the same
  // basis as the grid. Previously it stubbornly showed the full list.
  const listPrinters = getFlt();
  if(!P.length){el.innerHTML='<div class="empty-note">'+t('noDevicesList')+'</div>';return;}
  if(!listPrinters.length){el.innerHTML='<div class="empty-note">Kein Drucker passt zum Filter.</div>';return;}
  const alleAuf = listPrinters.length>0 && listPrinters.every(p=>LIST_AUF.has(p.ip));
  const sortiert = sortList(listPrinters);
  const th = (key, label, extra) => '<span class="sortbar' + (LIST_SORT.key===key?' aktiv':'') + (extra?' '+extra:'')
      + '" onclick="setListSort(\''+key+'\')">' + label + listSortPfeil(key) + '</span>';
  el.innerHTML='<div class="list-tools"><button class="hbtn btn-xs" id="btn-list-alle" onclick="listAlleUm()">'
      +(alleAuf?'▾ Alle einklappen':'▸ Alle ausklappen')+'</button>'
      +'<span class="note-sm">'+listPrinters.length+' von '+P.length+' Druckern</span></div>'
  +'<div class="list-thead"><span></span><span></span>'
      + th('modell','Modell') + th('name','Name') + th('online','Status') + th('datei','Datei')
      + th('fortschritt','Fortschritt') + th('restzeit','Restzeit','right') + th('ip','IP')
      + '</div>'
  +sortiert.map(p=>{
    const f=printerFields(p);
    const auf=LIST_AUF.has(p.ip);
    return '<div class="df-item zeile'+(f.fav?' fav':'')+(f.hatFehler?' hat-fehler':'')
        +(f.pausiert?' pausiert':'')+(auf?' auf':'')+'" id="li-'+esc(p.ip)+'" data-ip="'+esc(p.ip)
        +'" title="Klicken für Details" onclick="listToggle(this.dataset.ip, event)">'
      +'<span class="ll-pfeil">▸</span>'
      +feldHTML(f,'stern')+feldHTML(f,'modell')+feldHTML(f,'name')+feldHTML(f,'online')
      +feldHTML(f,'datei')+feldHTML(f,'fortschritt')+feldHTML(f,'restzeit')+feldHTML(f,'ip')
      +'</div>'
      +'<div class="list-detail'+(auf?' auf':'')+'" id="ld-'+esc(p.ip)+'"></div>';
  }).join('');
  listPrinters.forEach(p => { if (LIST_AUF.has(p.ip)) drawListRow(p.ip); });
  // Apply live updates immediately after render
  P.forEach(p => updateListRow(p.ip));
}



function qlSortWert(p, key){
  const st = STATUS[p.ip] || {};
  switch(key){
    case 'model':  return p.model || '';
    case 'name':   return p.name || '';
    case 'online': return st.online ? 1 : 0;
    case 'code':   return p.code ? 1 : 0;
    case 'serial': return p.serial || '';
    case 'ip':     return p.ip || '';
    default:       return p.name || '';
  }
}
function sortQL(list){
  const key=QL_SORT.key, dir=QL_SORT.dir;
  const zahl=(key==='online'||key==='code');
  return list.slice().sort((a,b)=>{
    const va=qlSortWert(a,key), vb=qlSortWert(b,key);
    let c = zahl ? (va-vb) : natCmp(va,vb);
    if(c===0) c=natCmp(a.name||'', b.name||'');
    return c*dir;
  });
}
function setQLSort(key){
  if(QL_SORT.key===key) QL_SORT.dir=-QL_SORT.dir; else QL_SORT={key:key,dir:1};
  renderQL();
}
function qlSortPfeil(key){ return QL_SORT.key!==key ? '' : '<span class="pf">'+(QL_SORT.dir>0?' ▲':' ▼')+'</span>'; }

function renderQL(){
  const el=document.getElementById('ql');
  setText('sc', '('+P.length+')');
  if(!P.length){el.innerHTML='<p class="note">'+t('noDevicesQL')+'</p>';return;}
  const th=(key,label)=>'<span class="sortbar'+(QL_SORT.key===key?' aktiv':'')+'" onclick="setQLSort(\''+key+'\')">'+label+qlSortPfeil(key)+'</span>';
  const kopf='<div class="ql-head"><span></span>'
    +th('model','Modell')+th('online','Status')+th('name','Name')+th('code','Zugangscode')
    +th('serial','Seriennummer')+th('ip','IP')+'<span></span></div>';
  el.innerHTML=kopf+sortQL(P).map((p,i)=>{const mc=modelColor(p.model);
    return '<div class="ql-row" id="qlrow-'+esc(p.ip)+'">'
      +'<span class="df-stern'+(p.fav?' on':'')+'" data-ip="'+esc(p.ip)+'" onclick="tileToggleFav(this.dataset.ip)">'+(p.fav?'★':'☆')+'</span>'
      +'<span class="mbadge" class="mbadge mbadge-color" style="--mc:'+mc+'">'+esc(p.model)+'</span>'
      +'<span class="ql-status ql-ping" title="'+esc(t('pingTitle'))+'" onclick="event.stopPropagation();pingPrinter(\''+esc(p.ip)+'\')">'+onlinePill(p.ip)+'</span>'
      +'<span class="ql-name"><span class="ql-name-txt">'+esc(p.name)+'</span>'+fwPillHTML(p.ip)+repairPill(p.ip)+'</span>'
      +'<span class="mono note-sm ql-code" title="Zugangscode — beim Bearbeiten sichtbar">'
        +(p.code?'•'.repeat(Math.min(8,p.code.length)):'—')+'</span>'
      +'<span class="mono note-sm ql-serial" title="Seriennummer">'+(p.serial?esc(p.serial):'—')+'</span>'
      +'<span class="mono note-sm ql-ip">'+esc(p.ip)+'</span>'
      +'<div class="ql-actions">'
        +qlRepairBtn(p)
        +qlCameraBtn(p)
        +qlBlinkBtn(p)
        +'<button class="hbtn btn-sm" onclick="editPrinter(\''+esc(p.ip)+'\')">Bearbeiten</button>'
        +'<button class="hbtn btn-sm red" onclick="delP(\''+esc(p.ip)+'\')">Löschen</button>'
      +'</div>'
      +'</div>';
  }).join('');
}

// editPrinter turns the row into a form. Access code and name can
// be changed this way without deleting and re-adding the printer.
function editPrinter(ip) {
  const p = P.find(x => x.ip === ip);
  const row = document.getElementById('qlrow-' + ip);
  if (!p || !row) return;
  row.classList.add('editing');
  // The same column order as without editing: star, model, name,
  // access code, serial number, IP. Previously the IP slipped onto its own row.
  row.innerHTML = '<span class="df-stern' + (p.fav ? ' on' : '') + '">' + (p.fav ? '★' : '☆') + '</span>'
    + '<input id="ed-model-' + esc(ip) + '" class="ed-model" value="' + esc(p.model) + '" placeholder="Modell" oninput="this.value=this.value.toUpperCase()">'
    + '<input id="ed-name-' + esc(ip) + '" class="ed-name" value="' + esc(p.name) + '" placeholder="Name">'
    + '<span></span>'
    + '<input id="ed-code-' + esc(ip) + '" class="ed-code mono" value="' + esc(p.code || '') + '" placeholder="Zugangscode">'
    + '<input id="ed-serial-' + esc(ip) + '" class="ed-serial mono" value="' + esc(p.serial || '') + '" placeholder="Seriennummer">'
    + '<input id="ed-ip-' + esc(ip) + '" class="ed-ip mono" value="' + esc(p.ip) + '" placeholder="IP-Adresse">'
    + '<div class="ql-actions">'
    +   '<button class="hbtn green btn-sm" onclick="savePrinter(\'' + esc(ip) + '\')">Speichern</button>'
    +   '<button class="hbtn btn-sm" onclick="renderQL()">Abbrechen</button>'
    + '</div>';
  const n = document.getElementById('ed-name-' + ip);
  if (n) { n.focus(); n.select(); }
  row.querySelectorAll('input').forEach(i => i.addEventListener('keydown', e => {
    if (e.key === 'Enter') savePrinter(ip);
    if (e.key === 'Escape') renderQL();
  }));
}

// pingPrinter checks a printer's reachability immediately (without waiting for the
// next status tick): /api/reach knocks on the MQTT/FTP/camera port.
// The verdict comes as a short message; the online pill is updated along.
async function pingPrinter(ip) {
  const p = P.find(x => x.ip === ip);
  if (!p) return;
  toast(t('pingLaeuft').replace('%s', p.name || ip), '');
  try {
    // Fetch fresh status so the MQTT row shows the REAL current error
    // (not up to 30 s old). IMPORTANT: do NOT reconnect here — that would
    // reset the error cause. Reconnect is a separate button.
    try { await loadStatus(); } catch(e) {}

    const r = await fetch('/api/reach?ip=' + encodeURIComponent(ip), {cache: 'no-store'});
    if (!r.ok) throw new Error('HTTP ' + r.status);
    const d = await r.json();
    const imNetz = (d.ports || []).some(x => x.offen);
    const hatSerial = !!(p.serial && String(p.serial).trim());

    let msg, art;
    if (!imNetz) { msg = t('pingOffline') + (d.urteil ? ' — ' + d.urteil : ''); art = 'er'; }
    else if (!hatSerial) { msg = t('pingKeineSerial'); art = 'er'; }
    else { msg = t('pingImNetz') + (d.urteil ? ' — ' + d.urteil : ''); art = 'ok'; }
    showPingResult(p, d, msg, art);
    return imNetz;
  } catch(e) { toast((t('pingFehler') || 'Ping fehlgeschlagen') + ': ' + (e.message || e), 'er'); }
}

// pingImNetzSuchen searches via network discovery (SSDP) for exactly this printer
// (by serial number). If it got a new IP via DHCP, it is
// adopted. If the search finds nothing, the device is off or on another
// network — then no tool can reach it.
async function pingImNetzSuchen(ip) {
  const p = P.find(x => x.ip === ip);
  if (!p) return;
  const serial = String(p.serial || '').trim();
  const btn = document.getElementById('ping-search-btn');
  if (btn) { btn.disabled = true; btn.textContent = t('pingRunning') || '…'; }
  toast(t('searchRunning') || 'Suche läuft …', '');
  try {
    const d = await (await fetch('/api/discover?timeout=6', {cache: 'no-store'})).json();
    const list = (d && d.printers) || [];
    const hit = serial ? list.find(x => x.serial === serial) : null;
    if (hit && hit.ip && hit.ip !== ip) {
      await applyNewIP(ip, hit.ip);
      toast((t('foundNewIp') || 'Neue Adresse übernommen: %s').replace('%s', hit.ip), 'ok');
      const o = document.getElementById('ping-overlay'); if (o) o.remove();
    } else if (hit) {
      toast(t('foundSameIp') || 'Gefunden unter derselben IP — reagiert aber nicht (Firmware/Netz).', 'er');
      if (btn) { btn.disabled = false; btn.textContent = '🔍 ' + t('pingSearchBtn'); }
    } else {
      toast(t('notFoundNet') || 'Im Netz nicht gefunden — Gerät aus, WLAN weg oder anderes Netz/VLAN.', 'er');
      if (btn) { btn.disabled = false; btn.textContent = '🔍 ' + t('pingSearchBtn'); }
    }
  } catch(e) {
    toast(String(e.message || e), 'er');
    if (btn) { btn.disabled = false; btn.textContent = '🔍 ' + t('pingSearchBtn'); }
  }
}

// pingReconnect triggers a fresh MQTT connection and updates the MQTT row
// in the open ping window over the next seconds.
async function pingReconnect(ip) {
  try {
    const btn = document.getElementById('ping-reconnect-btn');
    if (btn) { btn.disabled = true; btn.textContent = t('pingRunning') || '…'; }
    await fetch('/api/reconnect?ip=' + encodeURIComponent(ip), {method: 'POST'});
    toast(t('pingReconnectSent') || 'Neu verbinden ausgelöst', 'ok');
    [1500, 4000, 8000, 12000].forEach(ms => setTimeout(async () => {
      try { await loadStatus(); } catch(e) {}
      refreshPingMqtt(ip);
      const b = document.getElementById('ping-reconnect-btn');
      if (b && (STATUS[ip] && STATUS[ip].online)) { b.disabled = true; b.textContent = t('pingMqttOk'); }
    }, ms));
  } catch(e) { toast(String(e.message || e), 'er'); }
}

// mqttRow shows the real MQTT connection state including the error cause. Exactly that
// reveals why a printer stays "offline" despite an open port 8883 (e.g.
// "not authorized" = the access code is no longer valid).
function mqttRow(ip) {
  const s = STATUS[ip] || {};
  const on = !!s.online;
  const farbe = on ? 'var(--green)' : 'var(--red)';
  let txt = on ? t('pingMqttOk') : t('pingMqttOffline');
  if (!on && s.connect_err) txt += ' — ' + s.connect_err;
  return '<div class="ping-row" id="ping-mqtt-' + esc(ip) + '"><span class="ping-was">MQTT (Steuerung)</span>'
    + '<span style="color:' + farbe + '">' + (on ? '✓ ' : '✗ ') + esc(txt) + '</span></div>';
}

// refreshPingMqtt updates the MQTT row in the open ping window once
// the reconnect has taken effect.
function refreshPingMqtt(ip) {
  const zeile = document.getElementById('ping-mqtt-' + ip);
  if (!zeile) return; // Fenster zu
  const neu = mqttRow(ip);
  const tmp = document.createElement('div');
  tmp.innerHTML = neu;
  if (tmp.firstElementChild) zeile.replaceWith(tmp.firstElementChild);
}

// pingAuthFehler: true when the MQTT connection failed with a login error
// abgelehnt wurde (falscher Zugangscode).
function pingAuthFehler(ip) {
  const s = STATUS[ip] || {};
  return !s.online && /authoriz|not auth|password|passwort|bad user/i.test(String(s.connect_err || ''));
}

// codeAendern jumps to the "+ Printers" list and opens editing there
// for this printer — so the current access code is entered quickly.
function codeAendern(ip) {
  const o = document.getElementById('ping-overlay'); if (o) o.remove();
  try { sv('import'); } catch(e) {}
  setTimeout(() => {
    try {
      editPrinter(ip);
      const row = document.getElementById('qlrow-' + ip);
      if (row && row.scrollIntoView) row.scrollIntoView({block: 'center'});
    } catch(e) {}
  }, 150);
}

// showPingResult opens a window that lists the individual ports and
// stays OPEN until closed (the brief toast was gone too fast).
function showPingResult(p, d, msg, art) {
  const alt = document.getElementById('ping-overlay'); if (alt) alt.remove();
  const ports = (d.ports || []).slice().sort((a, b) => a.port - b.port);
  const zeile = x => {
    const farbe = x.offen ? 'var(--green)' : 'var(--red)';
    const status = x.offen ? t('pingPortOffen') : (x.grund || t('pingPortZu'));
    return '<div class="ping-row"><span class="ping-was">' + esc(x.was || '') + ' · Port ' + x.port + '</span>'
      + '<span style="color:' + farbe + '">' + (x.offen ? '✓ ' : '✗ ') + esc(status) + '</span></div>';
  };
  const html = '<div class="ping-overlay" id="ping-overlay" onclick="if(event.target===this)this.remove()">'
    + '<div class="modal-box ping-box">'
    + '<div class="ping-title">' + esc(d.name || p.name || p.ip) + ' — ' + esc(p.ip) + '<span class="dd-close" title="' + esc(t('close')) + '" onclick="var o=document.getElementById(\'ping-overlay\');if(o)o.remove()">✕</span></div>'
    + (msg ? '<div class="ping-msg ' + (art === 'er' ? 'rot' : 'gruen') + '">' + esc(msg) + '</div>' : '')
    + '<div class="ping-ports">' + ports.map(zeile).join('') + '</div>'
    + mqttRow(p.ip)
    + (pingAuthFehler(p.ip) ? '<div class="ping-hint">⚠ ' + t('pingAuthHint') + '</div>' : '')
    + (d.urteil ? '<div class="ping-urteil">' + esc(d.urteil) + '</div>' : '')
    + '<div class="btn-row" style="justify-content:flex-end;gap:8px;margin-top:14px">'
    + (pingAuthFehler(p.ip) ? '<button class="hbtn" onclick="codeAendern(\'' + esc(p.ip) + '\')">🔧 ' + t('pingCodeBtn') + '</button>' : '')
    + (p.serial ? '<button class="hbtn" id="ping-search-btn" onclick="pingImNetzSuchen(\'' + esc(p.ip) + '\')">🔍 ' + t('pingSearchBtn') + '</button>' : '')
    + (p.serial ? '<button class="hbtn primary" id="ping-reconnect-btn" onclick="pingReconnect(\'' + esc(p.ip) + '\')">🔄 ' + t('pingReconnectBtn') + '</button>' : '')
    + '<button class="hbtn" onclick="var o=document.getElementById(\'ping-overlay\');if(o)o.remove()">' + t('close') + '</button></div>'
    + '</div></div>';
  document.body.insertAdjacentHTML('beforeend', html);
}

async function savePrinter(ip) {
  const p = P.find(x => x.ip === ip);
  if (!p) return;
  const val = id => (document.getElementById(id + '-' + ip) || {}).value || '';
  const name = val('ed-name').trim(), code = val('ed-code').trim();
  const serial = val('ed-serial').trim(), model = val('ed-model').trim().toUpperCase();
  const neueIP = val('ed-ip').trim() || p.ip;
  // The name is enough. Access code and model are NOT needed to rename —
  // so a printer can be named without being connected to it or
  // having entered a code yet.
  if (!name) { toast('Name darf nicht leer sein', 'er'); return; }

  // The address changes more often than anything else in practice — a printer
  // gets a new one assigned by the network and then silently disappears.
  // Exactly that paralysed half a farm here. So it is editable,
  // but with a check.
  if (!/^(\d{1,3}\.){3}\d{1,3}$/.test(neueIP)) { toast('IP-Adresse sieht nicht richtig aus', 'er'); return; }
  if (neueIP.split('.').some(x => parseInt(x, 10) > 255)) { toast('IP-Adresse außerhalb des gültigen Bereichs', 'er'); return; }
  if (neueIP !== p.ip && P.some(x => x.ip === neueIP)) {
    toast('Diese IP gehört schon zu ' + P.find(x => x.ip === neueIP).name, 'er'); return;
  }

  const updated = {model, name, ip: neueIP, code, serial, fav: p.fav, added: p.added};
  try {
    const r = await fetch('/api/printers/' + encodeURIComponent(ip), {
      method: 'PUT', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(updated)
    });
    if (!r.ok) throw new Error((await r.text()).trim() || ('HTTP ' + r.status));
    const renamed = p.name !== name;
    const modelChanged = p.model !== model;
    const ipGeaendert = p.ip !== neueIP;
    Object.assign(p, updated);
    if (ipGeaendert) {
      // Everything tied to the old address is void.
      if (ST[ip]) { setTileState(ip, 'idle'); delete ST[ip]; }
      delete STATUS[ip];
      ST[neueIP] = 'idle';
      await loadPrinters();
      toast(name + ': Adresse auf ' + neueIP + ' geändert', 'ok');
    }
    render();                    // Liste, Kacheln, Sync-Auswahl
    updateStatusBar(neueIP);
    if (modelChanged) renderModelPills();
    if (renamed) {
      // The stream name derives from the printer name — the running
      // connection otherwise points at a name go2rtc no longer knows.
      if (ST[neueIP] === 'live') { stopWebRTC(neueIP); setTimeout(() => mountStreams([p]), 1200); }
      renderSyncPrinterList();
    }
    toast(name + ' gespeichert', 'ok');
    // Right after editing, check whether the printer (at a possibly new address)
    // is reachable — the server just rebuilt the MQTT connection.
    pingPrinter(neueIP);
  } catch(e) { toast(e.message || String(e), 'er'); }
}

function updStats(){
  setText('st0', P.length);
  const onlineCnt=P.filter(p=>STATUS[p.ip]&&STATUS[p.ip].online).length;
  setText('st-online', onlineCnt);
  setText('st-offline', P.length-onlineCnt);
  setText('st-paused', P.filter(p=>isPaused(p.ip)).length);
  setText('st-favs', P.filter(p=>p.fav).length);
  setText('st-fertig', P.filter(p=>istFertig(p.ip)).length);
  renderModelPills();
}

// ─── FARM-ALARME ──────────────────────────────────────────────────────────────
// Detects two events and reports them with sound + hint: (1) a print
// failed (gcode_state FAILED), (2) a printer goes offline MID-print.
// Per printer the last state is remembered so only the transition
// alarms and not anew on every status tick.
let ALARM_SOUND = (function(){ try { return localStorage.getItem('df_alarmSound') !== '0'; } catch(e){ return true; } })();
let ALARM_PREV = {};   // ip -> {state, online, printing}
let ALARM_ACTIVE = {}; // ip -> Meldung (aktuell anstehende Alarme)
let ALARM_AC = null;   // AudioContext (lazy)

function isPrintingState(s){ return s === 'RUNNING' || s === 'PREPARE' || s === 'SLICING'; }

function checkFarmAlarms() {
  const neu = [];
  P.forEach(p => {
    const s = STATUS[p.ip] || {};
    const cur = { state: (s.gcode_state||''), online: !!s.online, printing: isPrintingState(s.gcode_state||'') };
    const prev = ALARM_PREV[p.ip];
    if (prev) {
      // Print failed (new transition to FAILED)
      if (cur.state === 'FAILED' && prev.state !== 'FAILED') {
        neu.push({ ip: p.ip, kind: 'failed', text: (p.name||p.ip) + ': ' + t('alarmFailed') });
      }
      // Mitten im Druck offline gegangen
      else if (prev.online && prev.printing && !cur.online) {
        neu.push({ ip: p.ip, kind: 'offline', text: (p.name||p.ip) + ': ' + t('alarmOfflineMidPrint') });
      }
    }
    ALARM_PREV[p.ip] = cur;
    // Clear active alarms once the state normalizes.
    if (cur.state === 'FAILED' || (!cur.online && cur.printing)) {
      // bleibt aktiv
    } else {
      delete ALARM_ACTIVE[p.ip];
    }
  });
  neu.forEach(a => { ALARM_ACTIVE[a.ip] = a; });
  if (neu.length) {
    if (ALARM_SOUND) playAlarmBeep();
    neu.forEach(a => toast('⚠ ' + a.text, 'er'));
    notifyAlarm(neu);
  }
  updAlarmBadge();
}

function updAlarmBadge() {
  const n = Object.keys(ALARM_ACTIVE).length;
  setText('st-alarm', n);
  const pill = document.getElementById('pill-alarm');
  if (pill) pill.classList.toggle('alarm-live', n > 0);
  const bell = document.getElementById('alarm-bell');
  if (bell) bell.textContent = ALARM_SOUND ? '🔔' : '🔕';
}

function toggleAlarmSound() {
  ALARM_SOUND = !ALARM_SOUND;
  try { localStorage.setItem('df_alarmSound', ALARM_SOUND ? '1' : '0'); } catch(e){}
  if (ALARM_SOUND) {
    playAlarmBeep(); // kurzer Test-Ton bei Aktivierung
    // Ask once for permission for desktop notifications (user-triggered).
    try { if ('Notification' in window && Notification.permission === 'default') Notification.requestPermission(); } catch(e){}
  }
  updAlarmBadge();
  toast(t('alarmSound') + ': ' + (ALARM_SOUND ? t('on') : t('off')), 'ok');
}

// playAlarmBeep plays two short tones via the WebAudio API — no external
// audio file needed.
function playAlarmBeep() {
  try {
    if (!ALARM_AC) ALARM_AC = new (window.AudioContext || window.webkitAudioContext)();
    const ac = ALARM_AC;
    if (ac.state === 'suspended') ac.resume();
    [0, 0.28].forEach((delay, i) => {
      const osc = ac.createOscillator(), g = ac.createGain();
      osc.type = 'square';
      osc.frequency.value = i === 0 ? 880 : 660;
      g.gain.value = 0.0001;
      osc.connect(g); g.connect(ac.destination);
      const t0 = ac.currentTime + delay;
      g.gain.setValueAtTime(0.0001, t0);
      g.gain.exponentialRampToValueAtTime(0.25, t0 + 0.02);
      g.gain.exponentialRampToValueAtTime(0.0001, t0 + 0.22);
      osc.start(t0); osc.stop(t0 + 0.24);
    });
  } catch(e) {}
}

// notifyAlarm shows a desktop notification if the user has allowed it.
// Without permission nothing happens (no prompt spam).
function notifyAlarm(list) {
  try {
    if (!('Notification' in window)) return;
    if (Notification.permission === 'granted') {
      const body = list.map(a => a.text).join('\n');
      new Notification('Druckerfarm', { body: body });
    }
  } catch(e) {}
}
// The paused counter is also recomputed regularly. It depends on the
// print state, not a user action — without this tick a
// freshly paused machine would stay invisible until the next redraw.
setInterval(()=>{ setText('st-paused', P.filter(p=>isPaused(p.ip)).length); setText('st-fertig', P.filter(p=>istFertig(p.ip)).length); }, 4000);
function renderModelPills(){
  const el=document.getElementById('model-pills');if(!el)return;
  const seen={};P.forEach(p=>seen[p.model]=(seen[p.model]||0)+1);
  const active=[...AF][0]||null;
  el.innerHTML=Object.entries(seen).sort().map(([m,cnt])=>{const col=modelColor(m);const ia=active===m;
    return '<div class="stat-pill'+(ia?' active-model':'')+'" style="--mc:'+col+'" onclick="setPillFilter(\''+m+'\')">'
      +'<div class="sl">'+esc(m)+'</div><div class="sv cm">'+cnt+'</div></div>';
  }).join('');
}
// The YAML preview is gone; the path is a sub-item under go2rtc.
function renderYaml(){
  const el=document.getElementById('yout');if(!el)return;
  if(!P.length){el.innerHTML='<span class="muted">'+t('yamlEmpty')+'</span>';return;}
  let h='<span class="ycm"># go2rtc.yaml\n\n</span>';
  h+='<span class="yk">api:</span>\n  <span class="yk">origin:</span> <span class="yv">\'*\'</span>\n\n<span class="yk">streams:</span>\n';
  P.forEach(p=>{h+='  <span class="yk">'+esc(sname(p))+':</span>\n    - <span class="yv">rtspx://bblp:'+esc(p.code)+'@'+esc(p.ip)+':322/streaming/live/1</span>\n';});
  el.innerHTML=h;
}

// ─── ACTIONS ──────────────────────────────────────────────────────────────────
async function addM(){
  const model=document.getElementById('mm').value.trim().toUpperCase();
  if(!model){toast(t('errModel'),'er');return;}
  const name=document.getElementById('mn').value.trim(),ip=document.getElementById('mi').value.trim(),code=document.getElementById('mc').value.trim(),serial=document.getElementById('ms').value.trim();
  if(!name||!ip||!code||!serial){toast(t('errFields'),'er');return;}
  if(!/^(\d{1,3}\.){3}\d{1,3}$/.test(ip)){toast(t('errIp'),'er');return;}
  const kamAus=!!((document.getElementById('add-kamera-aus')||{}).checked);
  const blkAus=!!((document.getElementById('add-blink-aus')||{}).checked);
  const inRep=!!((document.getElementById('add-repair')||{}).checked);
  try{const p=await apiAdd({model,name,ip,code,serial,fav:false,added:Date.now(),kamera_aus:kamAus,blink_aus:blkAus,in_repair:inRep});
    if(kamAus)KAMERA_AUS[ip]=true;
    if(blkAus)BLINK_AUS[ip]=true;
    if(inRep){STATUS[ip]=STATUS[ip]||{};STATUS[ip].reparatur={in_repair:true,note:'',by:'',since:''};}
    P.push(p);ST[p.ip]='idle';render();
    ['mm','mn','mi','mc','ms'].forEach(id=>document.getElementById(id).value='');
    ['add-kamera-aus','add-blink-aus','add-repair'].forEach(id=>{const el=document.getElementById(id);if(el)el.checked=false;});
    toast(name+t('toastAdded'),'ok');}
  catch(e){toast(e.message,'er');}
}
async function delP(ip){const p=P.find(x=>x.ip===ip);if(!p)return;if(!confirm(t('confirmDel')))return;try{await apiDelete(ip);P=P.filter(x=>x.ip!==ip);delete ST[ip];render();toast(p.name+t('toastRemoved'),'er');}catch(e){toast(e.message,'er');}}
function confirmDelAll(){if(confirm(t('confirmDelAll')))delAll();}
async function delAll(){for(const p of[...P])await apiDelete(p.ip);P=[];ST={};render();toast(t('toastAllDeleted'),'er');}
async function impCSV(raw){
  raw = raw || '';
  if(!raw.trim()){toast(t('toastNoContent'),'er');return;}
  // Validate CSV format before sending
  const lines = raw.trim().split('\n').filter(l=>l.trim()&&!l.trim().startsWith('#'));
  if(lines.length===0){toast('Keine gültigen Zeilen gefunden','er');return;}
  // Detect separator: , or ;
  const sep = lines[0].includes(';') ? ';' : ',';
  const invalidLines=[];
  lines.forEach((line,i)=>{
    const parts=line.split(sep).map(s=>s.trim());
    if(parts.length<4){invalidLines.push('Zeile '+(i+1)+': zu wenig Felder ('+parts.length+'/4): '+line.substring(0,40));}
  });
  if(invalidLines.length>0){
    const ir=document.getElementById('ir');
    if(ir){setShown(ir, true);ir.className='ir er';ir.textContent='Format-Fehler: '+invalidLines[0];}
    toast('CSV Format-Fehler — erwartet: Name,IP,Code,Modell,Serial','er');
    return;
  }
  try{
    const res=await apiImport(raw);
    await loadPrinters();
    const ir=document.getElementById('ir');
    if(ir){setShown(ir, true);
      if(res.added>0){ir.className='ir ok';ir.textContent='✓ '+res.added+t('importOk')+(res.skipped>0?' ('+res.skipped+t('importSkipped')+')':'');toast(res.added+t('toastImported'),'ok');}
      else{ir.className='ir er';ir.textContent=t('importNone')+res.skipped+' Fehler/Duplikate)';}}
  }
  catch(e){
    const ir=document.getElementById('ir');
    if(ir){setShown(ir, true);ir.className='ir er';ir.textContent='Fehler: '+e.message;}
    toast('Import fehlgeschlagen: '+e.message,'er');
  }
}
async function dlYaml(){if(!P.length){toast(t('toastNoPrinters'),'er');return;}const a=document.createElement('a');a.href='/api/yaml?download=1';a.download='go2rtc.yaml';a.click();toast(t('toastYamlDl'),'ok');}
function expCSV(){if(!P.length){toast(t('toastNoPrinters'),'er');return;}window.location.href='/api/export';}


// ─── SYNC ─────────────────────────────────────────────────────────────────────
let syncPollTimer = null;

async function loadSyncConfig() {
  try {
    const r = await fetch('/api/sync/config');
    const cfg = await r.json();
    setVal('sync-ref-path', cfg.ref_path || '');
    setVal('sync-sd-path', cfg.sd_path  || '/');
    const sa = byId('sync-auto'); if (sa) sa.checked = !!cfg.auto_sync;
    setVal('sync-interval', cfg.auto_interval || 60);
    const mode = cfg.mode || (cfg.auto_sync ? 'auto' : 'manual');
    const rb = document.querySelector('input[name="sync-mode"][value="'+mode+'"]'); if (rb) rb.checked = true;
    syncModeUI(mode);
    renderSyncModelFolders(cfg.model_folders);
  } catch(e) {}
}

// syncModeUI shows/hides the appropriate areas depending on the mode.
function syncModeUI(mode){
  const auto = document.getElementById('sync-auto-cfg');
  const dz = document.getElementById('syncDropZone');
  if (auto) auto.style.display = (mode==='auto'||mode==='both') ? '' : 'none';
  if (dz) dz.style.display = (mode==='manual'||mode==='both') ? '' : 'none';
}
function aktuellerSyncModus(){
  const r = document.querySelector('input[name="sync-mode"]:checked');
  return r ? r.value : 'manual';
}
function setSyncMode(mode){
  syncModeUI(mode);
  saveSyncConfig();
}

// renderSyncModelFolders baut je vorhandenem Druckermodell ein Ordner-Feld.
function renderSyncModelFolders(map){
  map = map || {};
  const el = document.getElementById('sync-model-folders'); if (!el) return;
  const models = [...new Set((P||[]).map(p => String(p.model||'').toUpperCase()).filter(Boolean))].sort();
  if (!models.length){ el.innerHTML = '<span class="note-sm">'+t('syncNoModels')+'</span>'; return; }
  el.innerHTML = models.map(m =>
    '<div class="mf-row"><span class="mf-model">'+esc(m)+'</span>'
    +'<input type="text" class="repo-input mf-input" data-model="'+esc(m)+'" value="'+esc(map[m]||'')+'" placeholder="C:\\gcode\\'+esc(m)+'"></div>'
  ).join('');
}
function sammleModelFolders(){
  const out = {};
  document.querySelectorAll('#sync-model-folders .mf-input').forEach(i => {
    const v = (i.value||'').trim(); if (v) out[i.dataset.model] = v;
  });
  return out;
}
async function checkSyncDiff(){
  const hint = document.getElementById('sync-diff-hint'); if (!hint) return;
  hint.style.display = 'block';
  hint.innerHTML = '<span class="note-sm">'+t('syncChecking')+'</span>';
  try{
    const rows = await (await fetch('/api/sync/diff')).json() || [];
    if (!rows.length){ hint.innerHTML = '<span class="note-sm">'+t('syncNoDiff')+'</span>'; return; }
    hint.innerHTML = '<div class="mf-diff-title">⚠ '+t('syncDiffTitle')+'</div>'
      + rows.map(x => '<div class="mf-diff-row"><b>'+esc(x.name)+'</b> ('+esc(x.model)+') — '
        + (x.err ? esc(x.err) : (x.extra||[]).map(esc).join(', ')) + '</div>').join('');
  }catch(e){ hint.innerHTML = '<span class="note-sm">'+esc(e.message||String(e))+'</span>'; }
}

async function saveSyncConfig() {
  const cfg = {
    ref_path:      (byId('sync-ref-path') || {}).value?.trim() || '',
    sd_path:       (byId('sync-sd-path') || {}).value?.trim() || '/',
    auto_sync:     !!(byId('sync-auto') || {}).checked,
    auto_interval: parseInt((byId('sync-interval')||{}).value) || 60,
    mode:          aktuellerSyncModus(),
    model_folders: sammleModelFolders(),
  };
  try {
    await fetch('/api/sync/config', {
      method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify(cfg)
    });
    toast('Sync-Einstellungen gespeichert ✓','ok');
  } catch(e) { toast(e.message,'er'); }
}

async function syncAction(action) {
  try {
    await fetch('/api/sync/'+action, {method:'POST'});
    pollSyncStatus();
  } catch(e) { toast(e.message,'er'); }
}

async function pollSyncStatus() {
  try {
    const r = await fetch('/api/sync/status');
    const s = await r.json();
    updateSyncUI(s);
    if (s.status === 'running' || s.status === 'paused') {
      syncPollTimer = setTimeout(pollSyncStatus, 800);
    }
  } catch(e) {}
}

function updateSyncUI(s) {
  const labels = {idle:'IDLE',running:'RUNNING',paused:'PAUSED'};
  const badge = byId('sync-status-badge');
  if (badge) {
    badge.textContent = labels[s.status] || String(s.status||'').toUpperCase();
    badge.className = 'sd-badge sd-badge-' + (s.status||'idle');
  }
  

  const progDiv = document.getElementById('sync-progress');
  if (s.status !== 'idle') {
    setShown(progDiv, true);
    const pct = s.total > 0 ? Math.round((s.done/s.total)*100) : 0;
    const pb = byId('sync-progress-bar'); if (pb) pb.style.setProperty('--w', pct+'%');
    setText('sync-progress-lbl', s.current ? s.current + ' ('+s.done+'/'+s.total+')' : s.done+'/'+s.total+' Drucker');
    setText('sync-file-lbl', s.current_file ? '↑ '+s.current_file : '');
  } else {
    setShown(progDiv, false);
  }

  setText('sync-uploaded', s.uploaded || 0);
  setText('sync-skipped', s.skipped  || 0);
  const errBadge = document.getElementById('sync-errors-badge');
  if (s.errors && s.errors.length > 0) {
    setShown(errBadge, true);
    setText('sync-error-count', s.errors.length);
  } else {
    setShown(errBadge, false);
  }

  if (s.last_sync) setText('sync-last', 'Letzter Sync: '+s.last_sync);

  // Log
  if (s.log && s.log.length) {
    setText('sync-log', s.log.join('\n'));
    const logEl = document.getElementById('sync-log');
    logEl.scrollTop = logEl.scrollHeight;
  }

  // Button states
  setDis('sync-btn-start', s.status === 'running');
  setDis('sync-btn-pause', s.status !== 'running');
  setDis('sync-btn-stop', s.status === 'idle');
}



// ─── SYNC UPLOAD DROP ZONE ────────────────────────────────────────────────────
let syncDropFiles = []; // Array of File objects
// syncModelFilter remembers the last chosen model even across a
// page switch or restart (localStorage). Previously the variable stayed
// but initSyncUpload always marked the "All" pill active,
// while the list was still filtered by the old model — pill and
// tatsaechlicher Filter liefen auseinander.
let syncModelFilter = (function(){ try { return localStorage.getItem('df_syncModelFilter') || 'all'; } catch(e){ return 'all'; } })();

function initSyncUpload() {
  // Build model filter pills
  const models = [...new Set(P.map(p => p.model).filter(Boolean))].sort();
  const bar = document.getElementById('sync-model-filter');
  if (!bar) return;
  // Keep the remembered model only if there is still a printer for it.
  if (syncModelFilter !== 'all' && !models.includes(syncModelFilter)) syncModelFilter = 'all';
  const act = (m) => syncModelFilter === m ? ' active' : '';
  let html = '<span class="note-sm">Modell:</span>';
  html += '<span class="sync-model-pill'+act('all')+'" data-model="all" onclick="setSyncModel(\'all\')">Alle</span>';
  models.forEach(m => {
    const mc = modelColor(m);
    html += '<span class="sync-model-pill'+act(m)+'" data-model="'+esc(m)+'" onclick="setSyncModel(\''+esc(m)+'\')" style="--mc:'+mc+'">'+esc(m)+'</span>';
  });
  bar.innerHTML = html;

  // Build printer list
  renderSyncPrinterList();

  // Drop zone events
  const dz = document.getElementById('syncDropZone');
  if (!dz) return;
  dz.addEventListener('dragover', e => { e.preventDefault(); dz.classList.add('drag-over'); });
  dz.addEventListener('dragleave', () => dz.classList.remove('drag-over'));
  dz.addEventListener('drop', e => {
    e.preventDefault();
    dz.classList.remove('drag-over');
    const alle = [...e.dataTransfer.files];
    const files = alle.filter(f => isAllowedPrintFile(f.name));
    const abgelehnt = alle.filter(f => !isAllowedPrintFile(f.name));
    if (abgelehnt.length) toast(t('wrongFormat') + ' ' + abgelehnt.map(f => f.name).join(', '), 'er');
    addSyncFiles(files);
  });

  const fi = document.getElementById('syncFileInput');
  if (fi) fi.addEventListener('change', e => {
    const alle = [...e.target.files];
    const files = alle.filter(f => isAllowedPrintFile(f.name));
    const abgelehnt = alle.filter(f => !isAllowedPrintFile(f.name));
    if (abgelehnt.length) toast(t('wrongFormat') + ' ' + abgelehnt.map(f => f.name).join(', '), 'er');
    addSyncFiles(files);
    e.target.value = '';
  });
}

function setSyncModel(model) {
  syncModelFilter = model;
  try { localStorage.setItem('df_syncModelFilter', model); } catch(e){}
  document.querySelectorAll('.sync-model-pill').forEach(el => {
    el.classList.toggle('active', el.dataset.model === model);
  });
  renderSyncPrinterList();
}

function syncSortWert(p, key) {
  const st = STATUS[p.ip] || {};
  if (key === 'model') return p.model || '';
  if (key === 'online') return st.online ? 1 : 0;
  return p.name || '';
}
function sortSyncList(list) {
  const key = SYNC_SORT.key, dir = SYNC_SORT.dir;
  const zahl = (key === 'online');
  return list.slice().sort((a, b) => {
    const va = syncSortWert(a, key), vb = syncSortWert(b, key);
    let c = zahl ? (va - vb) : natCmp(va, vb);
    if (c === 0) c = natCmp(a.name || '', b.name || '');
    return c * dir;
  });
}
function setSyncSort(key) {
  if (SYNC_SORT.key === key) SYNC_SORT.dir = -SYNC_SORT.dir; else SYNC_SORT = {key: key, dir: 1};
  renderSyncPrinterList();
}
function syncSortPfeil(key) { return SYNC_SORT.key !== key ? '' : '<span class="pf">' + (SYNC_SORT.dir > 0 ? ' ▲' : ' ▼') + '</span>'; }

function renderSyncPrinterList() {
  const el = document.getElementById('syncPrinterList');
  if (!el) return;
  const gefiltert = syncModelFilter === 'all' ? P : P.filter(p => p.model === syncModelFilter);
  const filtered = sortSyncList(gefiltert);
  const th = (key, label) => '<span class="sortbar' + (SYNC_SORT.key===key?' aktiv':'')
      + '" onclick="setSyncSort(\''+key+'\')">' + label + syncSortPfeil(key) + '</span>';
  const kopf = '<div class="sync-sortbar">' + th('name','Name') + th('model','Modell') + th('online','Status') + '</div>';
  el.innerHTML = kopf + filtered.map((p, i) => {
    const mc = modelColor(p.model);
    const idx = P.indexOf(p);
    return '<div class="sync-printer-item" id="srow-'+esc(p.ip)+'" data-sdip="'+esc(p.ip)+'" data-sdi="'+idx+'">'
      + '<input type="checkbox" class="sync-cb" data-ip="'+esc(p.ip)+'" onchange="updateSyncSelCount()">'
      + '<div class="mdot" class="sync-dot" style="--mc:'+mc+'"></div>'
      + '<span class="sync-printer-label">'+esc(p.name)+'</span>'
      + onlinePill(p.ip)
      + '<span class="mono-xs mr6">'+esc(p.model)+'</span>'
      + '<span class="sd-file-count" id="sdcount2-'+esc(p.ip)+'"></span>'
      + '<span class="mono-xs ml4" id="sdsize2-'+esc(p.ip)+'"></span>'
      + '<button class="sd-dl-btn ml6" id="sdbtn-'+esc(p.ip)+'" data-idx="'+idx+'" data-ip="'+esc(p.ip)+'" onclick="loadSDFiles2(parseInt(this.dataset.idx),this.dataset.ip)">↺ '+t('loadData')+'</button>'
      + '<span class="sd-hitcount" id="sdhits-'+esc(p.ip)+'"></span>'
      + '</div>'
      + '<div class="sd-search" id="sdsearch-'+esc(p.ip)+'"></div>'
      + '<div class="sd-files" id="sdfiles2-'+esc(p.ip)+'"></div>';
  }).join('');
  // The selection survives the list rebuild. Previously it was gone as soon as
  // one left the view and re-entered it — in the middle of the
  // Upload besonders aergerlich.
  document.querySelectorAll('.sync-cb').forEach(cb => { cb.checked = SYNC_SEL.has(cb.dataset.ip); });
  updateSyncSelCount();
  renderUploadProgress();
}

// loadAllSDFiles loads the file lists of the selected printers. Two bugs
// were here: it ran over ALL rows instead of the selection, and
// loadSDFiles2, as an accordion, collapsed every other row again — in the end
// exactly one printer was open, no matter how many had been loaded.
async function loadAllSDFiles() {
  const sel = syncSelectedPrinters();
  const btn = document.getElementById('btn-loadall');
  if (!sel.length) {
    toast('Drucker wählen', 'er');
    return;
  }
  if (btn) { btn.disabled = true; btn.innerHTML = '<span class="spinner spinner-sm"></span> lädt …'; }
  busyStart('lädt Dateilisten …');
  let done = 0;
  try {
    for (const p of sel) {
      const row = document.querySelector('[data-sdip="' + CSS.escape(p.ip) + '"]');
      if (!row) continue;
      await loadSDFiles2(parseInt(row.dataset.sdi), p.ip, {keepOthers: true, forceOpen: true});
      done++;
      if (btn) btn.innerHTML = '<span class="spinner spinner-sm"></span> ' + done + ' / ' + sel.length;
    }
  } finally {
    busyEnd();
    if (btn) { btn.disabled = false; btn.textContent = '↺ Inhalte laden'; }
  }
  toast(done + ' Drucker geladen', 'ok');
}

// SD_KIND merkt sich je Drucker, welche Dateiart in der geladenen Liste gezeigt
// wird: 'print' (Druckdateien), 'video' (Zeitraffer) oder 'snapshot'.
let SD_KIND = {};

// setSDKind schaltet die angezeigte Dateiart um und lädt die Liste neu.
function setSDKind(ip, kind) {
  SD_KIND[ip] = kind;
  loadSDFiles2(0, ip, {keepOthers: true, forceOpen: true});
}

// sdKindSeg baut den Umschalter Druckdateien / Videos / Snapshots.
function sdKindSeg(ip) {
  const cur = SD_KIND[ip] || 'print';
  const b = (k, label) => '<button class="sd-kind-btn'+(cur===k?' on':'')+'" data-ip="'+esc(ip)+'" data-k="'+k
    + '" onclick="event.stopPropagation();setSDKind(this.dataset.ip,this.dataset.k)">'+label+'</button>';
  return '<div class="sd-kind-seg">'
    + b('print', t('sdKindPrint')) + b('video', t('sdKindVideo')) + b('snapshot', t('sdKindSnap'))
    + '</div>';
}

async function loadSDFiles2(i, ip, opts) {
  opts = opts || {};
  const files = document.getElementById('sdfiles2-'+ip);
  const count = document.getElementById('sdcount2-'+ip);
  if (!files) return;

  const isOpen = files.classList.contains('open');
  if (!opts.keepOthers) {
    // A single click stays an accordion: one list open, the rest closed.
    document.querySelectorAll('[id^="sdfiles2-"].open').forEach(el => el.classList.remove('open'));
    if (isOpen && !opts.forceOpen) return;
  }

  const kind = SD_KIND[ip] || 'print';
  files.classList.add('open');
  files.innerHTML = sdKindSeg(ip) + '<div class="sd-loading"><span class="spinner spinner-sm"></span> '+t('mediaLoading')+'</div>';

  const one = document.getElementById('sdbtn-'+ip);
  if (one) { one.disabled = true; one.innerHTML = '<span class="spinner spinner-sm"></span>'; }
  try {
    // Druckdateien über /api/sync/sdlist, Medien über /api/media/list.
    const url = kind === 'print'
      ? '/api/sync/sdlist?ip='+encodeURIComponent(ip)
      : '/api/media/list?ip='+encodeURIComponent(ip)+'&kind='+kind;
    const r = await fetch(url);
    const d = await r.json();

    if (d.error) {
      files.innerHTML = sdKindSeg(ip) + '<div class="sd-file-row red">✗ '+esc(ftpErrText(d))+'</div>';
      return;
    }
    const items = (d.files || []).map(f => ({ name: f.name || f.path, size: f.size||0, path: f.path||'', modified: f.modified||0 }));
    if (count) count.textContent = items.length + ' ' + t('mediaFiles');
    const sizeEl = document.getElementById('sdsize2-'+ip);
    const totalSize = items.reduce((n,f)=>n+(f.size||0),0);
    if (sizeEl) sizeEl.textContent = totalSize ? '· '+fmtSize(totalSize) : '';
    // Freien Speicher anzeigen.
    fetch('/api/sync/diskspace?ip='+encodeURIComponent(ip))
      .then(r=>r.json()).then(ds=>{
        if (sizeEl && (ds.free||ds.total)) {
          let info = '';
          if (ds.total) info += fmtSize(ds.total) + ' total';
          if (ds.free)  info += ' · ' + fmtSize(ds.free) + ' ' + t('mediaFree');
          sizeEl.textContent = (totalSize ? '📦 '+fmtSize(totalSize) : '') + (info ? '  💾 '+info : '');
        }
      }).catch(()=>{});

    if (!items.length) {
      files.innerHTML = sdKindSeg(ip) + '<div class="sd-file-row muted">'+t('mediaEmpty')+'</div>';
      return;
    }

    const isPrint = kind === 'print';
    files.innerHTML = sdKindSeg(ip)
      + '<div class="sd-file-tools">'
        + '<label class="sd-check"><input type="checkbox" class="sd-all" data-ip="'+esc(ip)+'" onclick="event.stopPropagation();sdAlleUm(this.dataset.ip,this.checked)"> alle</label>'
        + '<span class="sd-sel-info" id="sdselinfo-'+esc(ip)+'"></span>'
        + '<button class="hbtn btn-xs btn-danger push-right" id="sddel-'+esc(ip)+'" data-ip="'+esc(ip)+'" onclick="event.stopPropagation();sdDeleteSelected(this.dataset.ip)" disabled>'+t('mediaDelete')+'</button>'
      + '</div>'
      + items.slice().sort((a,b)=>natCmp(a.name,b.name)).map(f => {
        const kb = fmtSize(f.size);
        const dateStr = f.modified ? new Date(f.modified*1000).toLocaleDateString('de-CH',{day:'2-digit',month:'2-digit',year:'2-digit',hour:'2-digit',minute:'2-digit'}) : '';
        const ref = isPrint ? f.name : f.path;
        const dl = isPrint
          ? "dlSDFile(this.dataset.ip,this.dataset.ref)"
          : "dlMedia2(this.dataset.ip,'"+kind+"',this.dataset.ref)";
        const del = isPrint
          ? "delSDFile(this.dataset.ip,this.dataset.ref,this)"
          : "delMedia2(this.dataset.ip,'"+kind+"',this.dataset.ref,this)";
        return '<div class="sd-file-row" data-file="'+esc(ref)+'">'
          + '<input type="checkbox" class="sd-fcb" data-ip="'+esc(ip)+'" data-file="'+esc(ref)+'" onclick="event.stopPropagation();sdAuswahl(this.dataset.ip)">'
          + '<span class="sd-file-name" title="'+esc(f.name)+'">'+esc(f.name)+'</span>'
          + (dateStr ? '<span class="sep-note">'+dateStr+'</span>' : '')
          + '<span class="sd-file-size">'+kb+'</span>'
          + '<button class="sd-dl-btn" data-ip="'+esc(ip)+'" data-ref="'+esc(ref)+'" onclick="event.stopPropagation();'+dl+'">⬇</button>'
          + '<button class="sd-dl-btn btn-danger ml2" data-ip="'+esc(ip)+'" data-ref="'+esc(ref)+'" onclick="event.stopPropagation();'+del+'">✕</button>'
          + '</div>';
      }).join('');
  } catch(e) {
    files.innerHTML = sdKindSeg(ip) + '<div class="sd-file-row red">'+esc(e.message)+'</div>';
  } finally {
    if (one) { one.disabled = false; one.textContent = '↺ '+t('loadData'); }
  }
}

// dlMedia2 lädt eine Medien-Datei (Video/Snapshot) vom Drucker herunter.
function dlMedia2(ip, kind, path) {
  const a = document.createElement('a');
  a.href = '/api/media/download?ip='+encodeURIComponent(ip)+'&kind='+kind+'&path='+encodeURIComponent(path);
  a.click();
  toast('⬇ '+path.split('/').pop(), 'ok');
}

// delMedia2 löscht eine einzelne Medien-Datei vom Drucker.
async function delMedia2(ip, kind, path, btn) {
  if (!confirm(t('mediaDelConfirm'))) return;
  btn.disabled = true; btn.textContent = '⟳';
  try {
    const r = await fetch('/api/media/delete?ip='+encodeURIComponent(ip)+'&kind='+kind+'&path='+encodeURIComponent(path), {method:'DELETE'});
    const d = await r.json();
    if (d.error) throw new Error(ftpErrText(d));
    btn.closest('.sd-file-row').remove();
    toast(t('mediaDeleted'), 'ok');
  } catch(e) {
    btn.disabled = false; btn.textContent = '✕';
    toast(e.message, 'er');
  }
}

function syncToggleAll() {
  const allChecked = [...document.querySelectorAll('.sync-cb')].every(cb => cb.checked);
  document.querySelectorAll('.sync-cb').forEach(cb => cb.checked = !allChecked);
  updateSyncSelCount();
}

let SYNC_SEL = new Set();
function updateSyncSelCount() {
  SYNC_SEL = new Set([...document.querySelectorAll('.sync-cb:checked')].map(cb => cb.dataset.ip));
  const checked = document.querySelectorAll('.sync-cb:checked').length;
  const total = document.querySelectorAll('.sync-cb').length;
  const el = document.getElementById('sync-sel-count');
  if (el) el.textContent = checked + ' / ' + total + ' ' + t('selected');
  const btn = document.getElementById('syncUploadBtn');
  if (btn) btn.disabled = uploadRunning() || checked === 0 || syncDropFiles.length === 0;
  const tb = document.getElementById('sync-toggle-btn');
  if (tb) tb.textContent = (checked === total ? '☑ Alle abwählen' : '☑ Alle auswählen');
}

function addSyncFiles(files) {
  files.forEach(f => {
    if (!syncDropFiles.find(x => x.name === f.name)) syncDropFiles.push(f);
  });
  renderDropList();
  updateSyncSelCount();
}

function removeDropFile(name) {
  syncDropFiles = syncDropFiles.filter(f => f.name !== name);
  renderDropList();
  updateSyncSelCount();
}

function renderDropList() {
  const dz = document.getElementById('syncDropZone');
  const list = document.getElementById('syncDropList');
  const txt = document.getElementById('syncDropText');
  if (!list) return;
  if (syncDropFiles.length === 0) {
    dz.classList.remove('has-files');
    setShown(txt, true);
    list.innerHTML = '';
    return;
  }
  dz.classList.add('has-files');
  setShown(txt, false);
  list.innerHTML = '<div class="ok-note">'+syncDropFiles.length+' Datei(en) bereit</div>'
    + syncDropFiles.map(f =>
      '<div class="drop-file-item">'
      + '<span class="drop-file-name">'+esc(f.name)+'</span>'
      + '<span class="note-xs mx8">'+fmtSize(f.size)+'</span>'
      + '<span class="drop-file-rm" data-name="'+esc(f.name)+'" onclick="event.stopPropagation();removeDropFile(this.dataset.name)">✕</span>'
      + '</div>'
    ).join('');
}

// ─── UPLOAD ALS VORGANG ───────────────────────────────────────────────────────
//
// The upload used to run as a plain loop in the click handler: if one left the
// view, the progress lines remained but were gone after a
// rebuild, and pausing or cancelling was impossible. Now
// the state lives in UPLOAD and the display is drawn from it.

let UPLOAD = null;

function uploadRunning() { return !!(UPLOAD && UPLOAD.active); }

function toggleUploadPause() {
  if (!uploadRunning()) return;
  UPLOAD.paused = !UPLOAD.paused;
  const b = document.getElementById('syncPauseBtn');
  if (b) b.textContent = UPLOAD.paused ? '▶ Fortsetzen' : '⏸ Pause';
  if (UPLOAD.paused) UPLOAD.note = 'angehalten';
  renderUploadProgress();
}

function cancelUpload() {
  if (!uploadRunning()) return;
  UPLOAD.cancelled = true;
  UPLOAD.paused = false;
  UPLOAD.note = 'wird abgebrochen …';
  renderUploadProgress();
}

function waitWhilePaused() {
  return new Promise(res => {
    const tick = () => {
      if (!UPLOAD || !UPLOAD.paused || UPLOAD.cancelled) return res();
      setTimeout(tick, 200);
    };
    tick();
  });
}

// renderUploadProgress draws the progress display entirely from the
// state. So it looks the same after a view switch as before.
function renderUploadProgress() {
  const progDiv = document.getElementById('syncUploadProgress');
  const pauseB  = document.getElementById('syncPauseBtn');
  const cancelB = document.getElementById('syncCancelBtn');
  if (!progDiv) return;
  const run = uploadRunning();
  if (pauseB)  pauseB.classList.toggle('hidden', !run);
  if (cancelB) cancelB.classList.toggle('hidden', !run);
  if (!UPLOAD) { setShown(progDiv, false); progDiv.innerHTML = ''; return; }
  setShown(progDiv, true);
  progDiv.innerHTML = UPLOAD.ips.map(ip => {
    const p  = P.find(x => x.ip === ip);
    const st = UPLOAD.status[ip] || {text: 'Warte …', pct: 0, failed: false};
    return '<div class="sync-progress-item">'
      + '<span class="sync-pi-name">' + (p ? esc(p.name) : esc(ip)) + '</span>'
      + '<div class="sync-pi-bar"><div class="sync-pi-fill bar-fill' + (st.failed ? ' failed' : '')
      + '" style="--w:' + st.pct + '%"></div></div>'
      + '<span class="sync-pi-status">' + esc(st.text) + '</span>'
      + '</div>';
  }).join('');
  setText('syncUploadStatus', UPLOAD.note || (UPLOAD.done + ' / ' + UPLOAD.ips.length));
  if (typeof updateJobsBadge === 'function') updateJobsBadge();
  if (JOBS_OFFEN && typeof renderJobsPanel === 'function') renderJobsPanel();
}

function isAllowedPrintFile(name){
  return /\.(gcode|3mf)$/i.test(String(name||''));
}
async function startDropSync() {
  if (uploadRunning()) return;
  const selectedIPs = [...document.querySelectorAll('.sync-cb:checked')].map(cb => cb.dataset.ip);
  if (!selectedIPs.length) { toast('Drucker wählen', 'er'); return; }
  if (!syncDropFiles.length) { toast('Keine Dateien ausgewählt', 'er'); return; }

  const files = syncDropFiles.slice();
  UPLOAD = {active: true, paused: false, cancelled: false, ips: selectedIPs,
            status: {}, done: 0, note: '', total: selectedIPs.length};
  selectedIPs.forEach(ip => { UPLOAD.status[ip] = {text: 'Warte …', pct: 0, failed: false}; });

  const btn = document.getElementById('syncUploadBtn');
  if (btn) btn.disabled = true;
  const pb = document.getElementById('syncPauseBtn');
  if (pb) pb.textContent = '⏸ Pause';
  renderUploadProgress();
  busyStart('lädt hoch …');

  try {
    for (const ip of selectedIPs) {
      if (UPLOAD.cancelled) break;
      let errors = 0;
      for (let fi = 0; fi < files.length; fi++) {
        await waitWhilePaused();
        if (UPLOAD.cancelled) break;
        const file = files[fi];
        UPLOAD.status[ip] = {text: '⟳ ' + file.name, pct: fi / files.length * 100, failed: false};
        renderUploadProgress();
        const formData = new FormData();
        formData.append('file', file);
        try {
          const r = await fetch('/api/sync/upload?ip=' + encodeURIComponent(ip), {method: 'POST', body: formData});
          const res = await r.json();
          if (!r.ok || res.error) throw new Error(res.error || 'Fehler');
        } catch(e) {
          errors++;
          UPLOAD.status[ip] = {text: '✗ ' + (e.message || 'Fehler'), pct: (fi + 1) / files.length * 100, failed: true};
        }
        if (!errors) UPLOAD.status[ip].pct = (fi + 1) / files.length * 100;
        renderUploadProgress();
      }
      if (UPLOAD.cancelled) { UPLOAD.status[ip] = {text: 'abgebrochen', pct: 0, failed: true}; break; }
      if (!errors) UPLOAD.status[ip] = {text: '✓ ' + files.length + ' Datei(en)', pct: 100, failed: false};
      UPLOAD.done++;
      renderUploadProgress();
    }
  } finally {
    busyEnd();
    const cancelled = UPLOAD.cancelled;
    UPLOAD.active = false;
    UPLOAD.note = cancelled ? '✕ abgebrochen nach ' + UPLOAD.done + ' Drucker(n)'
                            : '✓ ' + UPLOAD.done + ' Drucker fertig';
    renderUploadProgress();
    if (btn) btn.disabled = false;
    loadUploadLog();
    toast(cancelled ? 'Upload abgebrochen' : 'Upload abgeschlossen ✓', cancelled ? 'er' : 'ok');
  }
}

// ─── SD CARD BROWSER ──────────────────────────────────────────────────────────
let sdData = {};  // ip -> {state:'idle'|'loading'|'done'|'error', files:[], file_count, total_size, error}

function fmtSize(bytes) {
  if (!bytes) return '0 B';
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024*1024) return (bytes/1024).toFixed(1) + ' KB';
  if (bytes < 1024*1024*1024) return (bytes/1024/1024).toFixed(1) + ' MB';
  return (bytes/1024/1024/1024).toFixed(2) + ' GB';
}

function loadSDList() {
  renderSyncPrinterList();
}

function updateSDSummary() {
  let totalFiles = 0, totalSize = 0, totalPrinters = 0;
  P.forEach(p => {
    const d = sdData[p.ip];
    if (d && d.state === 'done') {
      totalPrinters++;
      totalFiles += d.file_count||0;
      totalSize  += d.total_size||0;
    }
  });
  if (totalPrinters > 0) {
    setText('sd-total-printers', totalPrinters);
    setText('sd-total-files', totalFiles);
    setText('sd-total-size', fmtSize(totalSize));
    setShown(document.getElementById('sd-summary'), true);
  }
}

async function sdToggle(i) {
  const p = P[i];
  if (!p) return;
  const filesEl = document.getElementById('sdfiles-'+i);
  const row     = document.getElementById('sdrow-'+i);
  const arrow   = document.getElementById('sdarr-'+i);
  if (!filesEl) return;

  const isOpen = filesEl.classList.contains('open');

  // Close all others
  document.querySelectorAll('.sd-printer-row.open').forEach(r=>r.classList.remove('open'));
  document.querySelectorAll('.sd-files.open').forEach(f=>f.classList.remove('open'));
  document.querySelectorAll('.sd-arrow').forEach(a=>a.classList.remove('open'));

  if (!isOpen) {
    row.classList.add('open');
    filesEl.classList.add('open');
    if (arrow) arrow.classList.add('open');
    await sdLoad(i);
  }
}

async function sdLoad(i) {
  const p = P[i];
  const filesEl = document.getElementById('sdfiles-'+i);
  const d = sdData[p.ip];

  // Already loaded — just render
  if (d && d.state === 'done') { sdRenderFiles(i); return; }
  if (d && d.state === 'loading') return;

  // Mark loading
  sdData[p.ip] = {state:'loading'};
  const countEl = document.getElementById('sdcount-'+i);
  const errEl   = document.getElementById('sderr-'+i);
  if (countEl) countEl.textContent = '';
  if (errEl)   errEl.textContent = '';
  filesEl.innerHTML = '<div class="sd-loading">⟳ Verbinde mit '+esc(p.name)+' …</div>';

  try {
    const r = await fetch('/api/sync/sdlist?ip='+encodeURIComponent(p.ip));
    const detail = await r.json();
    if (detail.error) {
      const em = ftpErrText(detail);
      sdData[p.ip] = {state:'error', error: em};
      filesEl.innerHTML = '<div class="sd-file-row red">✗ '+esc(em)+'</div>';
      if (errEl) errEl.textContent = '✗ ' + em;
      return;
    }
    sdData[p.ip] = {state:'done', files: detail.files||[], file_count: detail.file_count||0, total_size: detail.total_size||0};
    if (countEl) countEl.textContent = (detail.file_count||0) + ' Dateien';
    if (document.getElementById('sdsize-'+i)) document.getElementById('sdsize-'+i).textContent = fmtSize(detail.total_size||0);
    sdRenderFiles(i);
    updateSDSummary();
  } catch(e) {
    sdData[p.ip] = {state:'error', error: e.message};
    filesEl.innerHTML = '<div class="sd-file-row red">✗ '+esc(e.message)+'</div>';
    if (errEl) errEl.textContent = '✗ ' + e.message;
  }
}

function sdRenderFiles(i) {
  const p = P[i];
  const filesEl = document.getElementById('sdfiles-'+i);
  const d = sdData[p.ip];
  if (!d || !d.files) return;
  if (!d.files.length) {
    filesEl.innerHTML = '<div class="sd-file-row muted">Keine .gcode Dateien</div>';
    return;
  }
  filesEl.innerHTML = d.files.slice().sort((a,b)=>natCmp(a.name,b.name)).map(f =>
    '<div class="sd-file-row">'
    + '<span class="sd-file-name" title="'+esc(f.name)+'">'+esc(f.name)+'</span>'
    + '<span class="sd-file-size">'+fmtSize(f.size)+'</span>'
    + '<button class="sd-dl-btn" data-ip="'+esc(p.ip)+'" data-file="'+esc(f.name)+'" onclick="event.stopPropagation();dlSDFile(this.dataset.ip,this.dataset.file)">⬇</button>'
    + '</div>'
  ).join('');
}

async function sdLoadAll() {
  for (let i = 0; i < P.length; i++) {
    const d = sdData[P[i].ip];
    if (!d || d.state === 'idle') await sdLoad(i);
  }
}


// ─── DATEIEN JE DRUCKER MARKIEREN UND LOESCHEN ────────────────────────────────
//
// Applies to a printer's directly loaded file list — independent of the
// search. Deletion goes via the same endpoint as single deletion, just
// in sequence for each marked file.
function sdCheckboxen(ip) {
  const wurzel = byId('sdfiles2-' + ip);
  return wurzel ? [...wurzel.querySelectorAll('.sd-fcb')] : [];
}

function sdAlleUm(ip, an) {
  sdCheckboxen(ip).forEach(cb => { cb.checked = !!an; });
  sdAuswahl(ip);
}

function sdAuswahl(ip) {
  const alle = sdCheckboxen(ip);
  const markiert = alle.filter(cb => cb.checked);
  const info = byId('sdselinfo-' + ip);
  if (info) info.textContent = markiert.length ? markiert.length + ' markiert' : '';
  const btn = byId('sddel-' + ip);
  if (btn) btn.disabled = markiert.length === 0;
  const all = byId('sdfiles2-' + ip) && byId('sdfiles2-' + ip).querySelector('.sd-all');
  if (all) all.checked = alle.length > 0 && markiert.length === alle.length;
}

async function sdDeleteSelected(ip) {
  const dateien = sdCheckboxen(ip).filter(cb => cb.checked).map(cb => cb.dataset.file);
  if (!dateien.length) return;
  const kind = SD_KIND[ip] || 'print';
  const name = (P.find(p => p.ip === ip) || {name: ip}).name;
  if (!confirm(dateien.length + ' Datei(en) von "' + name + '" löschen?\n\n'
      + dateien.slice(0, 12).map(x=>x.split('/').pop()).join('\n') + (dateien.length > 12 ? '\n… und ' + (dateien.length - 12) + ' weitere' : ''))) return;

  const btn = byId('sddel-' + ip);
  if (btn) { btn.disabled = true; btn.innerHTML = '<span class="spinner spinner-sm"></span> löscht …'; }
  busyStart('löscht Dateien …');
  let ok = 0;
  const fehler = [];
  try {
    for (const datei of dateien) {
      try {
        const url = kind === 'print'
          ? '/api/sync/delete?ip=' + encodeURIComponent(ip) + '&file=' + encodeURIComponent(datei)
          : '/api/media/delete?ip=' + encodeURIComponent(ip) + '&kind=' + kind + '&path=' + encodeURIComponent(datei);
        const r = await fetch(url, {method: 'DELETE'});
        const d = await r.json().catch(() => ({}));
        if (!r.ok || d.error) throw new Error(d.error || ('HTTP ' + r.status));
        const zeile = byId('sdfiles2-' + ip).querySelector('.sd-file-row[data-file="' + CSS.escape(datei) + '"]');
        if (zeile) zeile.remove();
        ok++;
      } catch(e) { fehler.push(datei + ': ' + (e.message || e)); }
    }
  } finally {
    busyEnd();
    if (btn) { btn.disabled = false; btn.textContent = 'Markierte löschen'; }
  }
  sdAuswahl(ip);
  toast(ok + ' gelöscht' + (fehler.length ? ', ' + fehler.length + ' fehlgeschlagen' : ''), fehler.length ? 'er' : 'ok');
  if (fehler.length) console.warn('Nicht gelöscht:', fehler);
}

async function delSDFile(ip, file, btn) {
  if (!confirm('Datei "' + file + '" von ' + (P.find(p=>p.ip===ip)||{name:ip}).name + ' löschen?')) return;
  btn.disabled = true;
  btn.textContent = '⟳';
  try {
    const r = await fetch('/api/sync/delete?ip='+encodeURIComponent(ip)+'&file='+encodeURIComponent(file), {method:'DELETE'});
    const d = await r.json();
    if (!r.ok) throw new Error(d.error || 'Fehler');
    btn.closest('.sd-file-row').remove();
    toast(file + ' gelöscht', 'ok');
  } catch(e) {
    btn.disabled = false;
    btn.textContent = '✕';
    toast('Fehler: ' + e.message, 'er');
  }
}

function dlSDFile(ip, file) {
  const a = document.createElement('a');
  a.href = '/api/sync/download?ip='+encodeURIComponent(ip)+'&file='+encodeURIComponent(file);
  a.download = file;
  a.click();
  toast('⬇ '+file+' wird heruntergeladen','ok');
}


// ─── CAMERA RESOLUTION ────────────────────────────────────────────────────────


// ─── SNAP MODE ────────────────────────────────────────────────────────────────
let SNAP_MODE = 2;        // 0 = video, N = snap interval in seconds (Vorgabe: 2s)
let SNAP_AUTO = false;    // wird erst wahr, wenn go2rtc und ffmpeg bestaetigt sind
let SNAP_TIMERS = {};     // ip -> timeout/interval id
let SNAP_URLS = {};       // ip -> blob URL currently shown (must be revoked)
let SNAP_INFLIGHT = {};   // ip -> true while a request is in flight
let SNAP_WARNED = false;  // report the failure reason only once per mode switch
let SNAP_SKIP = {};       // ip -> Zeitpunkt, ab dem wieder gefragt werden darf

function setSnapMode(seconds) {
  SNAP_MODE = seconds;
  SNAP_AUTO = true;
  SNAP_WARNED = false;
  const visible = new Set(getFlt().map(p => p.ip));
  // setTileState also clears a running video stream — previously it stayed
  // open when switching to 2s/5s/10s and could no longer be stopped.
  P.forEach(p => setTileState(p.ip, visible.has(p.ip) ? 'snap' : 'idle'));
  // A mode switch is an intentional restart: old pollers gone so they restart
  // with the new cadence (2/5/10 s). The additive startSnapAll would otherwise
  // skip the already-running ones and the old cadence would remain.
  stopSnapAll();
  markModeButton('btn-snap' + seconds);
  renderGrid();
}

function setVideoMode(res) {
  SNAP_MODE = 0;
  SNAP_AUTO = false;
  P.forEach(p => { if (ST[p.ip] === 'snap') setTileState(p.ip, 'idle'); });
  markModeButton('btn-' + res);
  setAllResolution(res); // Kameraaufloesung per MQTT
  renderGrid();
  toast(res + ' gesetzt — Streams mit ▶ starten', 'ok');
}

function markModeButton(id) {
  document.querySelectorAll('.res-btn').forEach(b => {
    b.classList.remove('res-active');
    b.classList.remove('working');
  });
  if (BUSY_AUS_TIMER) { clearTimeout(BUSY_AUS_TIMER); BUSY_AUS_TIMER = null; }
  BUSY_BTN = null;              // sonst haengt der Rahmen am alten Knopf fest
  const btn = document.getElementById(id);
  if (btn) btn.classList.add('res-active');
  paintBusy();
}

function startSnapAll(list) {
  if (SNAP_MODE <= 0) return;
  // Only tackle the tiles not yet running. Previously this
  // function first stopped ALL pollers and redistributed — if the grid was
  // redrawn multiple times at start, the staggering restarted each time and
  // the tiles at the end of the list never got their turn. Exactly why
  // some streams only started after a second click.
  const offen = (list || getFlt()).filter(p => ST[p.ip] === 'snap' && !SNAP_TIMERS[p.ip] && !SNAP_INFLIGHT[p.ip]
    && STATUS[p.ip] && STATUS[p.ip].online && !isRepair(p.ip) && !isCameraOff(p.ip)); // offline/Reparatur/Privat: kein Ladekreis
  if (!offen.length) return;
  const step = (SNAP_MODE * 1000) / Math.max(offen.length, 1);
  offen.forEach((p, i) => {
    SNAP_TIMERS[p.ip] = setTimeout(() => {
      delete SNAP_TIMERS[p.ip];
      if (SNAP_MODE > 0 && ST[p.ip] === 'snap') startSnapPoll(p.ip);
    }, Math.round(i * step));
  });
}

function stopSnapAll() {
  Object.values(SNAP_TIMERS).forEach(t => { clearTimeout(t); clearInterval(t); });
  SNAP_TIMERS = {};
  Object.keys(SNAP_URLS).forEach(releaseSnapUrl);
  SNAP_INFLIGHT = {};
}

function releaseSnapUrl(ip) {
  const u = SNAP_URLS[ip];
  if (u) { URL.revokeObjectURL(u); delete SNAP_URLS[ip]; }
}

function startSnapPoll(ip) {
  stopSnapPoll(ip);
  if (SNAP_MODE <= 0) return;
  const ring = byId('snapring-' + ip);
  if (ring) ring.style.setProperty('--takt', (SNAP_EFFECTIVE || SNAP_MODE) + 's');
  fetchSnap(ip); // immediate first snap
  SNAP_TIMERS[ip] = setInterval(() => fetchSnap(ip), SNAP_MODE * 1000);
}

function stopSnapPoll(ip) {
  if (SNAP_TIMERS[ip]) { clearTimeout(SNAP_TIMERS[ip]); clearInterval(SNAP_TIMERS[ip]); delete SNAP_TIMERS[ip]; }
  releaseSnapUrl(ip);
  delete SNAP_INFLIGHT[ip];
}

// The server raises the interval when the printer count requires it — otherwise
// go2rtc gets more image requests than it can handle and crashes.
let SNAP_EFFECTIVE = 0;
function noteEffectiveInterval(r) {
  const eff = parseInt(r.headers.get('X-Snapshot-Interval') || '0', 10);
  if (!eff || eff === SNAP_EFFECTIVE) return;
  SNAP_EFFECTIVE = eff;
  const throttled = r.headers.get('X-Snapshot-Throttled') === '1';
  // The hint is a tooltip on the active mode button. As text in the
  // bar it made it jump by a line on every change.
  const btn = document.querySelector('.res-btn.res-active');
  if (btn) {
    btn.title = throttled
      ? 'Tatsächlich alle ' + eff + ' s — bei ' + P.length
        + ' Druckern kann go2rtc nicht öfter Bilder liefern, ohne abzustürzen.'
      : '';
  }
  // Adjust the poll cadence, otherwise the requests run into nothing
  if (throttled && SNAP_MODE > 0 && eff > SNAP_MODE) restartSnapTimers(eff);
}

function restartSnapTimers(seconds) {
  Object.keys(SNAP_TIMERS).forEach(ip => {
    clearTimeout(SNAP_TIMERS[ip]); clearInterval(SNAP_TIMERS[ip]);
    SNAP_TIMERS[ip] = setInterval(() => fetchSnap(ip), seconds * 1000);
  });
}

function setSnapError(ip, msg) {
  const ts = document.getElementById('snapts-' + ip);
  if (isCamera(ip)) {
    // Camera unreachable (e.g. 503): NO spinner, just a quiet
    // short message. The ring is hidden until an image comes again.
    const ring = document.getElementById('snapring-' + ip);
    if (ring) ring.style.display = 'none';
    if (ts) { ts.textContent = t('camUnavailable'); ts.classList.add('red'); }
    return;
  }
  if (ts) { ts.textContent = '⚠ ' + msg; ts.classList.add('red'); }
}

async function fetchSnap(ip) {
  if (NETZ_PAUSE) return;
  if (SNAP_INFLIGHT[ip]) return; // previous request still running — don't pile up
  if (SNAP_SKIP[ip] && Date.now() < SNAP_SKIP[ip]) return; // pausierter Stream
  if (!document.getElementById('snapimg-' + ip)) return; // tile not on screen
  SNAP_INFLIGHT[ip] = true;
  busyStart('holt Bilder …');

  const ctrl = new AbortController();
  const timeout = setTimeout(() => ctrl.abort(), Math.max(8000, SNAP_MODE * 2000));
  try {
    // Served by our own backend, not go2rtc directly: it derives the stream name
    // from the printer (no second copy of the naming rules to drift out of sync),
    // caches the frame and caps how many transcodes run at once.
    const r = await fetch('/api/snapshot/' + encodeURIComponent(ip) + '?max_age=' + SNAP_MODE,
                          {signal: ctrl.signal, cache: 'no-store'});
    noteEffectiveInterval(r);
    // Camera temporarily without image: the server deliberately answers with 200 + header
    // (no 503 → no red console error). Show "unavailable", ring
    // ausblenden, Pause einhalten.
    if (r.headers.get('X-Snapshot-Unavailable') === '1') {
      setSnapError(ip, 'nicht verfügbar');
      const wait = parseInt(r.headers.get('X-Snapshot-Retry-In') || '0', 10);
      if (wait > 0) SNAP_SKIP[ip] = Date.now() + wait * 1000;
      return;
    }
    if (!r.ok) {
      const reason = (await r.text()).trim() || ('HTTP ' + r.status);
      if (!SNAP_WARNED) { SNAP_WARNED = true; toast('Snapshot: ' + reason, 'er'); }
      setSnapError(ip, reason);
      // The server pauses this stream — while so, do not even ask,
      // otherwise the requests keep running against a closed door.
      const wait = parseInt(r.headers.get('X-Snapshot-Retry-In') || '0', 10);
      if (wait > 0) SNAP_SKIP[ip] = Date.now() + wait * 1000;
      return;
    }
    delete SNAP_SKIP[ip];
    const blob = await r.blob();
    if (!blob.size) { setSnapError(ip, 'leeres Bild'); return; }

    const img = document.getElementById('snapimg-' + ip);
    if (!img) return;
    const url = URL.createObjectURL(blob);
    const prev = SNAP_URLS[ip];
    SNAP_URLS[ip] = url;
    const drop = () => { if (prev) URL.revokeObjectURL(prev); };
    // Important: also remove the fullscreen cell's spinner — otherwise it would
    // stay black, because this onload overrides the hideFsSpin handler set
    // in the cell.
    img.onload = () => { drop(); hideFsSpin(ip); };
    img.onerror = drop;
    setShown(img, true);
    img.src = url;

    SNAPS[ip] = {ts: new Date().toLocaleTimeString()};
    const ts = document.getElementById('snapts-' + ip);
    if (ts) { ts.textContent = SNAPS[ip].ts; ts.classList.remove('red'); }
    const ring0 = document.getElementById('snapring-' + ip);
    if (ring0) ring0.style.display = ''; // nach einem Fehler wieder anzeigen
  } catch(e) {
    if (e.name !== 'AbortError') setSnapError(ip, e.message || String(e));
  } finally {
    clearTimeout(timeout);
    delete SNAP_INFLIGHT[ip];
    busyEnd();
    // The mini ring runs continuously; its cycle time is tied to the current
    // cadence so it matches the next image.
    const ring = byId('snapring-' + ip);
    if (ring) ring.style.setProperty('--takt', (SNAP_EFFECTIVE || SNAP_MODE || 2) + 's');
  }
}

// ─── COLS / FILTER / NAV ──────────────────────────────────────────────────────
function setCols(v,persist){
  v=parseInt(v)||6;
  const g=document.getElementById('printerGrid'); if(g) g.style.setProperty('--cols', v);
  const l=document.getElementById('colLbl'); if(l) l.textContent=v+'×';
  const r=document.getElementById('colR'); if(r && parseInt(r.value)!==v) r.value=v;
  if(persist!==false) saveSettings({cols:v});
}

// ─── EINSTELLUNGEN ────────────────────────────────────────────────────────────

function setStartFilter(v) {
  if (v !== 'all' && v !== 'online' && v !== 'offline') v = 'all';
  START_FILTER = v;
  saveSettings({start_filter: v});
  toast(t('startFilterSaved') || 'Startfilter gespeichert', 'ok');
}

// ─── SENDESEITE: Einstellungen ────────────────────────────────────────────────
async function loadBroadcastCfg() {
  try {
    const c = await (await fetch('/api/broadcast/config', {cache:'no-store'})).json();
    const v = document.getElementById('bc-view'); if (v) v.value = c.view || 'all';
    const cl = document.getElementById('bc-click'); if (cl) cl.checked = !!c.click;
    const co = document.getElementById('bc-cols'); if (co) co.value = c.cols || 9;
    const u = document.getElementById('bc-url'); if (u) u.value = c.url || '';
  } catch(e) {}
}
async function saveBroadcastCfg() {
  const view = ((document.getElementById('bc-view')||{}).value) || 'all';
  const click = !!((document.getElementById('bc-click')||{}).checked);
  const cols = parseInt((document.getElementById('bc-cols')||{}).value) || 9;
  try {
    await fetch('/api/broadcast/config', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({view, click, cols})});
    toast(t('bcSaved') || 'Sendeseite gespeichert', 'ok');
  } catch(e) { toast(e.message || String(e), 'er'); }
}
function openBroadcast() {
  const u = ((document.getElementById('bc-url')||{}).value) || '/broadcast';
  window.open(u, '_blank');
}

// ─── JOB-/BRETT-PLANER ────────────────────────────────────────────────────────
function jpIstFrei(p){
  const s = STATUS[p.ip]; if(!s || !s.online) return false;
  const st = String(s.gcode_state||'').toLowerCase();
  return st==='' || st==='idle' || st==='finish';
}
function jpFmtH(h){ const d=Math.floor(h/24), r=Math.round((h-d*24)*10)/10; return (d>0?d+'d ':'')+r+'h'; }
// ─── BRETT-TABELLE (Planer-Nachschlag) ───────────────────────────────────────
let BOARD_TABLE = [];
async function loadBoardTable(){
  try{ const r=await fetch('/api/boardtable',{cache:'no-store'}); const j=await r.json(); BOARD_TABLE=Array.isArray(j)?j:[]; }
  catch(e){ BOARD_TABLE=[]; }
  renderBoardTable();
}
function renderBoardTable(){
  const tb=byId('bt-body'); if(!tb) return;
  tb.innerHTML = BOARD_TABLE.map((e,i)=>
     '<tr>'
    +'<td><input class="bt-in" data-i="'+i+'" data-k="model" value="'+esc(e.model||'')+'"></td>'
    +'<td><input class="bt-in" data-i="'+i+'" data-k="file" value="'+esc(e.file||'')+'"></td>'
    +'<td><input class="bt-in" type="number" min="0" data-i="'+i+'" data-k="per_plate" value="'+(e.per_plate||0)+'"></td>'
    +'<td><input class="bt-in" type="number" min="0" step="0.1" data-i="'+i+'" data-k="hours" value="'+(e.hours||0)+'"></td>'
    +'<td><button class="hbtn btn-xs" onclick="boardTableDelRow('+i+')">✕</button></td>'
    +'</tr>').join('');
  tb.querySelectorAll('.bt-in').forEach(inp=>inp.addEventListener('input', boardTableSync));
  // Refill the picker at the top
  const sel=byId('jp-pick'); if(sel){
    const cur=sel.value;
    sel.innerHTML='<option value="">'+t('jpPickNone')+'</option>'+BOARD_TABLE.map((e,i)=>
      '<option value="'+i+'">'+esc((e.model||'?')+' · '+(e.file||'?'))+'</option>').join('');
    sel.value=cur;
  }
}
function boardTableSync(ev){
  const el=ev.target, i=+el.dataset.i, k=el.dataset.k;
  if(!BOARD_TABLE[i]) return;
  BOARD_TABLE[i][k] = (k==='per_plate')?parseInt(el.value||'0',10):(k==='hours')?parseFloat(el.value||'0'):el.value;
}
function boardTableAddRow(){ BOARD_TABLE.push({model:'',file:'',per_plate:0,hours:0}); renderBoardTable(); }
function boardTableDelRow(i){ BOARD_TABLE.splice(i,1); renderBoardTable(); }
async function boardTableSave(){
  try{
    const r=await fetch('/api/boardtable',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(BOARD_TABLE)});
    if(!r.ok) throw new Error('HTTP '+r.status);
    await loadBoardTable();
    toast(t('btSaved'),'ok');
  }catch(e){ toast('Fehlgeschlagen','er'); }
}
// boardPick fills figures/board and time/board from the chosen table row.
function boardPick(){
  const sel=byId('jp-pick'); if(!sel||sel.value==='') return;
  const e=BOARD_TABLE[+sel.value]; if(!e) return;
  if(e.per_plate>0){ const pp=byId('jp-perplate'); if(pp) pp.value=e.per_plate; }
  if(e.hours>0){ const tt=byId('jp-time'); if(tt) tt.value=e.hours; }
}
function calcJobPlan(){
  const target=Math.max(1, parseInt((byId('jp-target')||{}).value)||0);
  const per=Math.max(1, parseInt((byId('jp-perplate')||{}).value)||1);
  const tph=Math.max(0, parseFloat((byId('jp-time')||{}).value)||0);
  const mode=((byId('jp-printers')||{}).value)||'free';
  const boards=Math.ceil(target/per);
  const figs=boards*per;
  let printers = (mode==='all') ? P.length : P.filter(jpIstFrei).length;
  printers=Math.max(1, printers);
  const rounds=Math.ceil(boards/printers);
  const totalH=rounds*tph;
  const el=byId('jp-result'); if(!el) return; el.style.display='block';
  const eta = totalH>0 ? new Date(Date.now()+totalH*3600*1000) : null;
  el.innerHTML =
      '<div class="jp-line"><b>'+boards+'</b> '+t('jpBoards')+' → '+figs+' '+t('jpFigures')+(figs>target?(' (+'+(figs-target)+' '+t('jpExtra')+')'):'')+'</div>'
    + '<div class="jp-line">'+printers+' '+t('jpUsed')+' · '+rounds+' '+t('jpRounds')+(tph>0?(' · ~'+jpFmtH(totalH)+' '+t('jpTotal')):'')+'</div>'
    + (eta?('<div class="jp-line">'+t('jpEta')+': '+eta.toLocaleString()+'</div>'):'');
}

// ─── CLOUDFLARE-TUNNEL: Einstellungen ─────────────────────────────────────────
let CF_POLL = null;
async function loadTunnelCfg() {
  try {
    const c = await (await fetch('/api/tunnel/config', {cache:'no-store'})).json();
    const m = document.getElementById('cf-mode'); if (m) m.value = c.mode || 'public';
    const cl = document.getElementById('cf-click'); if (cl) cl.checked = !!c.click;
    const need = document.getElementById('cf-need'); if (need) need.style.display = c.installed ? 'none' : 'block';
  } catch(e) {}
  refreshTunnelStatus();
}
async function refreshTunnelStatus() {
  try {
    const s = await (await fetch('/api/tunnel/status', {cache:'no-store'})).json();
    const st = document.getElementById('cf-status');
    if (st) st.textContent = s.running ? (s.url ? t('cfOnline') : t('cfStarting')) : (s.err ? s.err : t('cfOffline'));
    const u = document.getElementById('cf-url'); if (u) u.value = s.url || '';
    const need = document.getElementById('cf-need'); if (need) need.style.display = s.installed ? 'none' : 'block';
    const start = document.getElementById('cf-start'); const stop = document.getElementById('cf-stop');
    if (start) start.disabled = s.running || !s.installed;
    if (stop) stop.disabled = !s.running;
  } catch(e) {}
}
async function saveTunnelCfg() {
  const mode = ((document.getElementById('cf-mode')||{}).value) || 'public';
  const click = !!((document.getElementById('cf-click')||{}).checked);
  try {
    await fetch('/api/tunnel/config', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({mode, click})});
    toast(t('cfSaved'), 'ok'); refreshTunnelStatus();
  } catch(e) { toast(e.message || String(e), 'er'); }
}
async function installTunnel() {
  const st = document.getElementById('cf-inst-status'); if (st) st.textContent = t('cfInstalling');
  try {
    await fetch('/api/tunnel/install', {method:'POST'});
    const iv = setInterval(async () => {
      try {
        const p = await (await fetch('/api/components/progress', {cache:'no-store'})).json();
        if (p.component === 'cloudflared') {
          if (st) st.textContent = p.step;
          if (p.done) { clearInterval(iv); if (st) st.textContent = p.error || t('cfInstalled'); loadTunnelCfg(); }
        }
      } catch(e) {}
    }, 1000);
  } catch(e) { toast(e.message || String(e), 'er'); }
}
async function startTunnelUI() {
  busyStart(t('cfStarting'));
  try {
    const r = await fetch('/api/tunnel/start', {method:'POST'}); if (!r.ok) throw new Error(await r.text());
    if (!CF_POLL) CF_POLL = setInterval(refreshTunnelStatus, 2000);
    toast(t('cfStartedMsg'), 'ok');
  } catch(e) { toast(e.message || String(e), 'er'); } finally { busyEnd(); refreshTunnelStatus(); }
}
async function stopTunnelUI() {
  try {
    await fetch('/api/tunnel/stop', {method:'POST'}); if (CF_POLL) { clearInterval(CF_POLL); CF_POLL = null; }
    toast(t('cfStopped'), 'ok');
  } catch(e) { toast(e.message || String(e), 'er'); } finally { refreshTunnelStatus(); }
}
function copyTunnelUrl() {
  const u = ((document.getElementById('cf-url')||{}).value) || '';
  if (u && navigator.clipboard) navigator.clipboard.writeText(u);
  toast(u ? t('cfCopied') : t('cfNoUrl'), u ? 'ok' : 'er');
}

// ─── YOUTUBE-LIVE: Einstellungen ──────────────────────────────────────────────
let YT_POLL = null;
async function loadYouTubeCfg() {
  try {
    const c = await (await fetch('/api/youtube/config', {cache:'no-store'})).json();
    const r = document.getElementById('yt-res'); if (r) r.value = c.resolution || '1920x1080';
    const b = document.getElementById('yt-bitrate'); if (b) b.value = c.bitrate_k || 4500;
    const fp = document.getElementById('yt-fps'); if (fp) fp.value = c.fps || 15;
    const k = document.getElementById('yt-key'); if (k && c.has_key) k.placeholder = '•••••• (gespeichert)';
  } catch(e) {}
  refreshYouTubeStatus();
}
async function refreshYouTubeStatus() {
  try {
    const s = await (await fetch('/api/youtube/status', {cache:'no-store'})).json();
    const st = document.getElementById('yt-status');
    if (st) st.textContent = s.running ? (t('ytLive') + ' · ' + (s.tiles||0) + ' · ' + Math.floor((s.seconds||0)/60) + 'm') : (s.err ? s.err : t('ytOffline'));
    const start = document.getElementById('yt-start'); const stop = document.getElementById('yt-stop');
    if (start) start.disabled = s.running || !s.ffmpeg_present;
    if (stop) stop.disabled = !s.running;
  } catch(e) {}
}
async function saveYouTubeCfg() {
  const key = ((document.getElementById('yt-key')||{}).value) || '';
  const resolution = ((document.getElementById('yt-res')||{}).value) || '1920x1080';
  const bitrate_k = parseInt((document.getElementById('yt-bitrate')||{}).value) || 4500;
  const fps = parseInt((document.getElementById('yt-fps')||{}).value) || 15;
  const body = {resolution, bitrate_k, fps};
  if (key.trim()) body.key = key.trim();
  try {
    await fetch('/api/youtube/config', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify(body)});
    const k = document.getElementById('yt-key'); if (k) { k.value = ''; k.placeholder = '•••••• (gespeichert)'; }
    toast(t('ytSaved'), 'ok');
  } catch(e) { toast(e.message || String(e), 'er'); }
}
async function startYouTubeUI() {
  busyStart(t('ytStarting'));
  try {
    const r = await fetch('/api/youtube/start', {method:'POST'}); if (!r.ok) throw new Error(await r.text());
    if (!YT_POLL) YT_POLL = setInterval(refreshYouTubeStatus, 3000);
    toast(t('ytStartedMsg'), 'ok');
  } catch(e) { toast(e.message || String(e), 'er'); } finally { busyEnd(); refreshYouTubeStatus(); }
}
async function stopYouTubeUI() {
  try {
    await fetch('/api/youtube/stop', {method:'POST'}); if (YT_POLL) { clearInterval(YT_POLL); YT_POLL = null; }
    toast(t('ytStopped'), 'ok');
  } catch(e) { toast(e.message || String(e), 'er'); } finally { refreshYouTubeStatus(); }
}

// ─── KI-ASSISTENT: mehrere Konnektoren ────────────────────────────────────────
let AI_CONNS=[], AI_ACTIVE='';
const AI_PROV_LABEL={openai:'OpenAI',anthropic:'Anthropic',ollama:'Ollama'};
async function loadAIConfig(){
  // Prompts (global) aus /api/ai/config, Konnektoren separat.
  try{
    const c = await (await fetch('/api/ai/config', {cache:'no-store'})).json();
    const pe=byId('ai-prompt-explain'); if(pe) pe.value=c.prompt_explain||'';
    const pvv=byId('ai-prompt-vision'); if(pvv) pvv.value=c.prompt_vision||'';
  }catch(e){}
  await loadAIConnectors();
}
async function loadAIConnectors(){
  try{
    const d = await (await fetch('/api/ai/connectors', {cache:'no-store'})).json();
    AI_CONNS = d.connectors||[]; AI_ACTIVE = d.active_id||'';
  }catch(e){ AI_CONNS=[]; AI_ACTIVE=''; }
  renderAIConnectors();
}
function renderAIConnectors(){
  const el=byId('ai-conn-list'); if(!el) return;
  if(!AI_CONNS.length){ el.innerHTML='<div class="ai-conn-empty">'+t('aiNoConn')+'</div>'; }
  else{
    el.innerHTML = AI_CONNS.map(c=>{
      const act = c.id===AI_ACTIVE;
      const meta = (AI_PROV_LABEL[c.provider]||c.provider) + (c.model?(' · '+esc(c.model)):'')
        + (c.provider==='ollama' ? (c.url?(' · '+esc(c.url)):'') : (c.has_key?(' · '+t('aiKeySet')):(' · '+t('aiNoKey'))));
      return '<div class="ai-conn-row'+(act?' active':'')+'">'
        + '<label class="ai-conn-pick"><input type="radio" name="ai-active" '+(act?'checked':'')+' onchange="aiSetActive(\''+esc(c.id)+'\')"><span class="ai-conn-name">'+esc(c.name)+'</span></label>'
        + '<span class="ai-conn-meta">'+meta+'</span>'
        + '<span class="ai-conn-btns"><button class="pc-btn" onclick="aiEditConnector(\''+esc(c.id)+'\')">'+t('edit')+'</button>'
        + '<button class="pc-btn danger" onclick="aiDeleteConnector(\''+esc(c.id)+'\')">'+t('delete')+'</button></span>'
        + '</div>';
    }).join('');
  }
  const st=byId('ai-status');
  if(st){ const a=AI_CONNS.find(c=>c.id===AI_ACTIVE);
    st.textContent = a ? (a.name+' · '+(AI_PROV_LABEL[a.provider]||a.provider)+(a.model?(' · '+a.model):'')) : t('aiNoConn'); }
}
async function aiAddConnector(){
  const name=((byId('ai-new-name')||{}).value||'').trim();
  const provider=((byId('ai-new-provider')||{}).value)||'openai';
  if(!name){ toast(t('aiNameMissing'),'er'); return; }
  try{
    const r=await fetch('/api/ai/connectors',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name,provider})});
    if(!r.ok) throw new Error(await r.text());
    const d=await r.json();
    byId('ai-new-name').value='';
    await loadAIConnectors();
    aiEditConnector(d.id); // gleich zum Ausfüllen von Modell/Key öffnen
  }catch(e){ toast(e.message||String(e),'er'); }
}
function aiEditConnector(id){
  const c=AI_CONNS.find(x=>x.id===id); if(!c) return;
  byId('ai-edit-id').value=c.id;
  byId('ai-edit-name').value=c.name||'';
  byId('ai-edit-provider').value=c.provider||'openai';
  byId('ai-edit-model').value=c.model||'';
  byId('ai-edit-url').value=c.url||'';
  const k=byId('ai-edit-key'); if(k){ k.value=''; k.placeholder=c.has_key?'•••••• ('+t('aiKeySet')+')':'sk-…'; }
  byId('ai-conn-edit-row').style.display='';
  aiEditProviderChange();
  byId('ai-edit-name').focus();
}
function aiEditProviderChange(){
  const prov=((byId('ai-edit-provider')||{}).value)||'openai';
  const keyRow=byId('ai-edit-key-row'), urlRow=byId('ai-edit-url-row');
  if(keyRow) keyRow.style.display = (prov==='ollama')?'none':'';
  if(urlRow) urlRow.style.display = (prov==='ollama')?'':'none';
}
async function aiSaveConnector(){
  const id=((byId('ai-edit-id')||{}).value)||''; if(!id) return;
  const body={ name:((byId('ai-edit-name')||{}).value)||'', provider:((byId('ai-edit-provider')||{}).value)||'openai', model:((byId('ai-edit-model')||{}).value)||'', url:((byId('ai-edit-url')||{}).value)||'' };
  const k=((byId('ai-edit-key')||{}).value)||''; if(k.trim()) body.key=k.trim();
  try{
    const r=await fetch('/api/ai/connector?id='+encodeURIComponent(id),{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
    if(!r.ok) throw new Error(await r.text());
    toast(t('aiSaved'),'ok'); aiCancelEdit(); loadAIConnectors();
  }catch(e){ toast(e.message||String(e),'er'); }
}
function aiCancelEdit(){ const r=byId('ai-conn-edit-row'); if(r) r.style.display='none'; }
async function aiDeleteConnector(id){
  const c=AI_CONNS.find(x=>x.id===id); if(!c) return;
  if(!confirm(t('aiDelConfirm')+' "'+c.name+'"?')) return;
  try{
    const r=await fetch('/api/ai/connector?id='+encodeURIComponent(id),{method:'DELETE'});
    if(!r.ok) throw new Error(await r.text());
    aiCancelEdit(); loadAIConnectors();
  }catch(e){ toast(e.message||String(e),'er'); }
}
async function aiSetActive(id){
  try{
    const r=await fetch('/api/ai/active?id='+encodeURIComponent(id),{method:'POST'});
    if(!r.ok) throw new Error(await r.text());
    AI_ACTIVE=id; renderAIConnectors();
  }catch(e){ toast(e.message||String(e),'er'); }
}
async function saveAIPrompts(){
  const body={ prompt_explain:((byId('ai-prompt-explain')||{}).value)||'', prompt_vision:((byId('ai-prompt-vision')||{}).value)||'' };
  try{
    await fetch('/api/ai/config',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
    toast(t('aiSaved'),'ok');
  }catch(e){ toast(e.message||String(e),'er'); }
}
function buildAIButtons(p){
  return '<div class="dd-sep"></div>'
    + '<div class="pc-row"><button class="pc-btn" onclick="kiErklaereFehler(\''+esc(p.ip)+'\')">'+t('aiExplain')+'</button>'
    + '<button class="pc-btn" onclick="aiCheckImage(\''+esc(p.ip)+'\')">'+t('aiVision')+'</button></div>'
    + '<div id="ai-out-'+esc(p.ip)+'" class="ai-out" style="display:none;"></div>';
}
async function kiErklaereFehler(ip){
  const box=byId('ai-out-'+ip); if(box){ box.style.display='block'; box.textContent=t('aiThinking'); }
  try{
    const r=await fetch('/api/ai/explain',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({ip})});
    if(!r.ok) throw new Error(await r.text());
    const d=await r.json();
    if(box) box.textContent = d.keine_fehler ? t('aiNoError') : (d.text||'');
  }catch(e){ if(box) box.textContent=(e.message||String(e)); }
}
async function aiCheckImage(ip){
  const box=byId('ai-out-'+ip); if(box){ box.style.display='block'; box.textContent=t('aiThinking'); }
  try{
    const r=await fetch('/api/ai/vision',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({ip})});
    if(!r.ok) throw new Error(await r.text());
    const d=await r.json();
    if(box) box.textContent = d.text||'';
  }catch(e){ if(box) box.textContent=(e.message||String(e)); }
}

function setErrorBlink(on) {
  ERROR_BLINK = !!on;
  markSeg('blink-', on ? 'on' : 'off');
  saveSettings({error_blink: ERROR_BLINK});
  toast(on ? 'Sonderbeleuchtung eingeschaltet' : 'Sonderbeleuchtung aus — Licht wird zurückgesetzt', 'ok');
}

// setShown toggles the class instead of the style property — so the
// presentation is in the stylesheet and not in the script.
// Some elements are hidden by default via the stylesheet (.view,
// .sd-files, .ir). For those, removing "hidden" is not enough — they
// need a rule that makes them visible. Hence the "shown" class.
function setShown(el, on) {
  if (!el) return;
  el.classList.toggle('hidden', !on);
  el.classList.toggle('shown', !!on);
}

// Accessors that silently do nothing if the element does not (yet)
// exist. Parts of the UI are only drawn at runtime — without these
// helpers a missing element would abort the whole calling function.
function byId(id) { return document.getElementById(id); }

// Indicates that something is running. Without it the UI seems dead while
// waiting on printers in the background.
let BUSY = 0;
function busyStart(what) {
  BUSY++;
  BUSY_WHAT = what || BUSY_WHAT;
  paintBusy();
}
function busyEnd() {
  BUSY = Math.max(0, BUSY - 1);
  paintBusy();
}
// The hint is no longer text in the bar — which made it jump on
// every request. Instead a frame runs around the active mode button.
let BUSY_WHAT = '';
// The frame flickered because on EVERY request the class was removed and
// re-added immediately — restarting the animation from
// scratch each time. With 42 printers that happens every second. Now only
// what actually changes is touched; if the frame is already running, it keeps
// ununterbrochen weiter.
// The frame gets a linger time. Without it it flickered every second:
// each image request is its own busyStart/busyEnd pair, and between two
// requests the counter briefly drops to zero. With 30 requests that was 60
// class changes. Now it only turns off when for half a second
// truly nothing runs anymore — then the frame stays on during a series
// continuously and turns off once at the end.
let BUSY_BTN = null, BUSY_AUS_TIMER = null;
const BUSY_NACHLAUF = 500;

function paintBusy() {
  const soll = BUSY > 0 ? document.querySelector('.res-btn.res-active') : null;

  if (soll) {
    if (BUSY_AUS_TIMER) { clearTimeout(BUSY_AUS_TIMER); BUSY_AUS_TIMER = null; }
    if (soll !== BUSY_BTN) {
      if (BUSY_BTN) { BUSY_BTN.classList.remove('working'); BUSY_BTN.title = ''; }
      BUSY_BTN = soll;
      soll.classList.add('working');
    }
    soll.title = BUSY_WHAT || 'lädt …';
    return;
  }

  // Nothing left to do — but turn off only after the linger time.
  if (!BUSY_BTN || BUSY_AUS_TIMER) return;
  BUSY_AUS_TIMER = setTimeout(() => {
    BUSY_AUS_TIMER = null;
    if (BUSY > 0) return;                 // zwischenzeitlich doch wieder etwas
    if (BUSY_BTN) { BUSY_BTN.classList.remove('working'); BUSY_BTN.title = ''; }
    BUSY_BTN = null;
  }, BUSY_NACHLAUF);
}
function setText(id, v) { const e = byId(id); if (e) e.textContent = v; }
function setVal(id, v)  { const e = byId(id); if (e) e.value = v; }
function setCls(id, v)  { const e = byId(id); if (e) e.className = v; }
function setDis(id, v)  { const e = byId(id); if (e) e.disabled = !!v; }

function markSeg(prefix, value) {
  ['dark','light','de','en','on','off'].forEach(v => {
    const el = document.getElementById(prefix + v);
    if (el) el.classList.toggle('active', v === value);
  });
}

function setTheme(name, persist) {
  THEME = (name === 'dark') ? 'dark' : 'light';
  document.documentElement.setAttribute('data-theme', THEME);
  markSeg('theme-', THEME);
  if (persist !== false) {
    saveSettings({theme: THEME});
  }
}

let SETTINGS_TIMER = null;
function saveSettings(patch) {
  clearTimeout(SETTINGS_TIMER);
  SETTINGS_TIMER = setTimeout(() => {
    fetch('/api/settings', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify(patch)})
      .catch(() => {});
  }, 250);
}

// Settings come from the server, no longer from browser storage — which is
// gone with the Edge profile as soon as the profile is recreated.
async function loadSettings() {
  try {
    const d = await (await fetch('/api/settings', {cache:'no-store'})).json();
    setTheme(d.theme, false);
    LANG = ['de','en','zh','es'].includes(d.lang) ? d.lang : 'de';
    markSeg('lang-', LANG);
    enableTranslation();
    setCols(d.cols, false);
    ERROR_BLINK = d.error_blink !== false;
    markSeg('blink-', ERROR_BLINK ? 'on' : 'off');
    BLINK_MODELS = d.blink_models || [];
    const bm = document.getElementById('blink-models');
    if (bm) bm.textContent = (d.blink_models || []).join(', ') || '–';
    START_FILTER = (d.start_filter==='online'||d.start_filter==='offline') ? d.start_filter : 'all';
    BLINK_AUS = (d.blink_aus && typeof d.blink_aus==='object') ? d.blink_aus : {};
    KAMERA_AUS = (d.kamera_aus && typeof d.kamera_aus==='object') ? d.kamera_aus : {};
    MEDIA_ENFORCE = (d.timelapse_aus && typeof d.timelapse_aus==='object') ? d.timelapse_aus : {};
    const sf = document.getElementById('start-filter'); if (sf) sf.value = START_FILTER;
    if (d.filter_pillen && typeof d.filter_pillen === 'object') {
      ['online','offline','paused','favs','fertig'].forEach(k => { if (k in d.filter_pillen) FILTER_PILLE[k] = d.filter_pillen[k] !== false; });
    }
    applyFilterPills();
    if (!window.BROADCAST_OPTS) { loadBroadcastCfg(); loadTunnelCfg(); loadYouTubeCfg(); loadAIConfig(); }
    if (!d.sprache_gewaehlt) showLanguagePicker();
    return d;
  } catch(e) {
    setTheme('light', false);
    return {theme:'light', lang:'de', cols:6};
  }
}
// Which filter pills are visible — default: all. The choice is in the
// settings and stored server-side.
let FILTER_PILLE = {online:true, offline:true, paused:true, favs:true, fertig:true};

function applyFilterPills() {
  ['online','offline','paused','favs','fertig'].forEach(k => {
    const pill = byId('pill-' + (k === 'favs' ? 'favs' : k));
    if (pill) setShown(pill, FILTER_PILLE[k] !== false);
    const cb = byId('fpz-' + k);
    if (cb) cb.checked = FILTER_PILLE[k] !== false;
  });
}

function setFilterPill(k, an) {
  FILTER_PILLE[k] = !!an;
  // Hiding the currently active filter falls back to "All".
  if (!an) {
    const aktiv = (k==='online'&&ONLINE_ONLY)||(k==='offline'&&OFFLINE_ONLY)||(k==='paused'&&PAUSED_ONLY)||(k==='favs'&&FAVS_ONLY)||(k==='fertig'&&FERTIG_ONLY);
    if (aktiv) setPillFilter('all');
  }
  applyFilterPills();
  saveSettings({filter_pillen: FILTER_PILLE});
}

function setPillFilter(val){
  AF.clear();
  ONLINE_ONLY  = (val==='online');
  OFFLINE_ONLY = (val==='offline');
  PAUSED_ONLY  = (val==='paused');
  FAVS_ONLY    = (val==='favs');
  FERTIG_ONLY  = (val==='fertig');
  const known=['all','online','offline','paused','favs','fertig'];
  if(known.indexOf(val)<0) AF.add(val);
  ['pill-all','pill-online','pill-offline','pill-paused','pill-favs','pill-fertig'].forEach(id=>{
    const el=document.getElementById(id); if(el) el.classList.remove('active-all','active-online');
  });
  const active=document.getElementById('pill-'+val);
  if(active) active.classList.add(val==='online'?'active-online':'active-all');
  renderModelPills();
  ZULETZT_SICHTBAR = getFlt().map(p => p.ip).slice().sort().join(',');
  renderGrid();
  if(CURRENT_SUBVIEW==='list') renderList();
}

// syncFilterDisplay brings the highlighted pill in line with the ACTUALLY
// applied filter. Without it the overview showed "All" after a
// view switch but still filtered by the last chosen model.
function syncFilterDisplay(){
  let val='all';
  if(ONLINE_ONLY)val='online';
  else if(OFFLINE_ONLY)val='offline';
  else if(PAUSED_ONLY)val='paused';
  else if(FAVS_ONLY)val='favs';
  else if(FERTIG_ONLY)val='fertig';
  else if(AF.size>0)val=[...AF][0];
  ['pill-all','pill-online','pill-offline','pill-paused','pill-favs','pill-fertig'].forEach(id=>{
    const el=document.getElementById(id); if(el) el.classList.remove('active-all','active-online');
  });
  const known=['all','online','offline','paused','favs','fertig'];
  if(known.indexOf(val)>=0){
    const a=document.getElementById('pill-'+val);
    if(a) a.classList.add(val==='online'?'active-online':'active-all');
  }
  renderModelPills();
}
let PAUSED_ONLY=false;
let ZULETZT_SICHTBAR='';

// ─── NETZWERKVERKEHR ANHALTEN ─────────────────────────────────────────────────
//
// "Pause" is meant literally here: the UI stops asking, and
// the server drops its connections. All recurring queries are
// therefore collected here so they can be halted together.
const tickers = [];
let NETZ_PAUSE = false;

function showNetPause(an) {
  NETZ_PAUSE = an;
  const z = byId('netz-aus');
  if (z) z.classList.toggle('hidden', !an);
  document.body.classList.toggle('netz-pause', an);
  const b = byId('btn-netzpause');
  if (b) b.textContent = an ? '▶ Netzwerkverkehr wieder starten' : '⏸ Kompletten Netzwerkverkehr pausieren';
}

async function netzPauseUmschalten() {
  const soll = !NETZ_PAUSE;
  const b = byId('btn-netzpause');
  if (b) b.disabled = true;
  try {
    const r = await fetch('/api/network/pause', {
      method: 'POST', headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({pausiert: soll})
    });
    if (!r.ok) throw new Error((await r.text()).trim() || ('HTTP ' + r.status));
    await r.json();

    if (soll) {
      // First stop everything that would start again by itself.
      tickers.forEach(clearInterval);
      tickers.length = 0;
      stopAllMedia();
      if (typeof listImagesAllOff === 'function') listImagesAllOff();
      showNetPause(true);
      toast('Netzwerkverkehr angehalten — nichts wird mehr geladen', 'ok');
    } else {
      showNetPause(false);
      startTickers();
      // go2rtc needs a moment before it delivers images again.
      setTimeout(() => { loadStatus(); resumeMedia(); }, 1500);
      toast('Netzwerkverkehr wieder aufgenommen', 'ok');
    }
  } catch(e) {
    toast(e.message || String(e), 'er');
  } finally { if (b) b.disabled = false; }
}

function startTickers() {
  tickers.forEach(clearInterval);
  tickers.length = 0;
  tickers.push(setInterval(chkG2, 15000));
  tickers.push(setInterval(loadErrors, 5000));
  tickers.push(setInterval(loadStatus, 30000));
}
function flt(q){SQ=q.toLowerCase();renderGrid();if(CURRENT_SUBVIEW==='list')renderList();}
function sv(v){
  document.querySelectorAll('.view').forEach(e=>setShown(e, false));
  document.querySelectorAll('.tab').forEach(e=>e.classList.remove('active'));
  // 'grid' and 'list' are sub-views inside view-uebersicht
  const mainView = (v==='grid'||v==='list'||v==='uebersicht') ? 'uebersicht' : v;
  const viewEl = document.getElementById('view-'+mainView);
  if (viewEl) setShown(viewEl, true);
  const tabEl = document.getElementById('tab-'+mainView);
  if (tabEl) tabEl.classList.add('active');
  // Images run only in the overview and in fullscreen. When switching to
  // File Sync, Printers or Settings they are stopped and restarted on
  // returning.
  if (mainView!=='uebersicht') stopAllMedia();
  if (v==='grid'||v==='list') setView(v);
  if (v==='setup') renderYaml();
  if (v==='list') renderList();
  if (v==='uebersicht') { setView('grid'); resumeMedia(); }
  if (v==='sync'){loadSyncConfig();pollSyncStatus();loadSDList();initSyncUpload();loadUploadLog();}
  // Jobs is its own page now (no side slider). JOBS_OFFEN drives whether the
  // running-jobs list re-renders on the 3s poll.
  JOBS_OFFEN = (v==='jobs');
  if (v==='jobs'){ loadJobs(); renderJobsPanel(); loadPlan(); }
}

// ─── FILE DROP ────────────────────────────────────────────────────────────────
const dz=document.getElementById('dz');
dz.addEventListener('dragover',e=>{e.preventDefault();dz.classList.add('drag');});
dz.addEventListener('dragleave',()=>dz.classList.remove('drag'));
dz.addEventListener('drop',e=>{e.preventDefault();dz.classList.remove('drag');if(e.dataTransfer.files[0])rdFf(e.dataTransfer.files[0]);});
function rdF(inp){if(inp.files[0])rdFf(inp.files[0]);inp.value='';}
// The file is read and imported directly — the intermediate field for
// pasting no longer exists.
function rdFf(f){
  const r=new FileReader();
  r.onload=e=>{toast(f.name+t('toastFileLoaded'),'ok');impCSV(String(e.target.result||''));};
  r.readAsText(f);
}

// ─── HELPERS ──────────────────────────────────────────────────────────────────
function fbCp(t){const e=document.createElement('textarea');e.value=t;document.body.appendChild(e);e.select();document.execCommand('copy');document.body.removeChild(e);}
function toast(msg,type='ok'){const c=document.getElementById('toasts'),t=document.createElement('div');t.className='toast '+type;t.innerHTML='<span>'+(type==='ok'?'✓':'✕')+'</span> '+msg;c.appendChild(t);setTimeout(()=>t.remove(),3200);}

// ─── FULLSCREEN VIEWS ─────────────────────────────────────────────────────────
let G2_RECOVERY_STREAMS = new Set();
let G2_RESTARTING = false;

function fmtRemaining(sec) {
  if (!sec || sec <= 0) return '';
  const d = Math.floor(sec/86400), h = Math.floor((sec%86400)/3600), m = Math.floor((sec%3600)/60);
  return (d>0?d+'d ':'')+((h>0||d>0)?h+'h ':'')+m+'m';
}

let CURRENT_SUBVIEW = 'grid';
function setView(v) {
  CURRENT_SUBVIEW = v;
  document.body.classList.toggle('fs-mode', v==='fsv'||v==='fsc');
  ['grid','list','fsv','fsc'].forEach(id => {
    const el=document.getElementById('view-'+id); if(el) setShown(el, false);
    const tab=document.getElementById('tab-'+id); if(tab) tab.classList.remove('active');
  });
  const el=document.getElementById('view-'+v); if(el) setShown(el, true);
  const tab=document.getElementById('tab-'+v); if(tab) tab.classList.add('active');
  // The stream bar belongs to the grid. No images run in the list, so
  // there is no switch for it there.
  const sb=document.getElementById('stream-bar'); if(sb) sb.classList.toggle('hidden', v!=='grid');
  const tb=document.getElementById('top-bar'); if(tb) tb.classList.toggle('hidden', v==='fsv'||v==='fsc');
  if(v==='list') { renderList(); stopAllMedia(); }
  if(v!=='list') listImagesAllOff();
  // When switching back to the grid the images must restart.
  // stopAllMedia() turned SNAP_AUTO off — without resetting it,
  // renderGrid() had no effect and all tiles stayed dark.
  if(v==='grid') { if (SNAP_MODE > 0) SNAP_AUTO = true; syncFilterDisplay(); renderGrid(); }
  if(v==='fsv') renderFSV();
  if(v==='fsc') renderFSC();
}

// stopAllMedia ends every running connection and every snapshot query.
// Needed as soon as a view without images becomes visible — otherwise
// 42 Verbindungen im Hintergrund weiter, obwohl niemand hinsieht.
function stopAllMedia() {
  SNAP_AUTO = false;
  setTileStates(P, 'idle');
  if (typeof listImagesAllOff === 'function') listImagesAllOff();
}

// resumeMedia brings the images back when the overview is visible again.
function resumeMedia() {
  if (CURRENT_SUBVIEW !== 'grid') return;
  if (SNAP_MODE > 0) SNAP_AUTO = true; else return;
  setTileStates(getFlt(), activeMode());
  renderGrid();
}

function renderFSV() {
  const el = document.getElementById('view-fsv');
  if (!el) return;
  const printers = getFlt();
  const n = printers.length || 1;
  const cols = calcFSCols(n, 16/9);
  const rows = Math.ceil(n / cols);
  const cellH = Math.floor(window.innerHeight / rows);

  clearGridDOM();
  adoptForFullscreen(printers);

  el.className = 'view fs-view';
  el.innerHTML = '<button class="fs-esc" onclick="exitFS()">✕ ESC</button>'
    + '<div class="fs-grid" style="--cols:'+cols+';--cellh:'+cellH+'px">'
    + printers.map(p => fsvCell(p)).join('')
    + CAMS.map(c => fsvCamCell(c)).join('')
    + '</div>';

  requestAnimationFrame(() => { mountStreams(printers); mountCamStreams(); });
}

// fsvCamCell: fullscreen video cell for a camera — image + name only, no
// Fortschrittsbalken, keine Statusfarbe.
function fsvCamCell(cam){
  const id = cam.id;
  const inner = SNAP_MODE > 0
    ? '<img id="snapimg-'+esc(id)+'" src="" class="fill-media" onload="hideFsSpin(\''+esc(id)+'\')" onerror="this.classList.add(\'hidden\');">'
    : '<video id="vid-'+esc(id)+'" autoplay muted playsinline class="fill-media"></video>';
  return '<div class="scan-cell fsv-cell cam-cell" data-ip="'+esc(id)+'">'
    + '<div class="fsv-name">'+esc(cam.name)+'</div>'
    + inner
    + '<div class="fs-spin" id="fsspin-'+esc(id)+'"><div class="cam-spinner"></div></div>'
    + '</div>';
}

function calcFSCols(n, aspect) {
  // Find cols that best fills the screen with given aspect ratio per cell
  // aspect = width/height of each cell (16/9 for video)
  const W = window.innerWidth, H = window.innerHeight;
  let bestCols = 1, bestWaste = Infinity;
  for (let cols = 1; cols <= n; cols++) {
    const rows = Math.ceil(n / cols);
    const cellW = W / cols;
    const cellH = cellW / aspect;
    const usedH = cellH * rows;
    // Penalise if rows overflow screen
    const waste = usedH > H ? (usedH - H) * 10 : (H - usedH);
    if (waste < bestWaste) { bestWaste = waste; bestCols = cols; }
  }
  return bestCols;
}

function renderFSC() {
  const el = document.getElementById('view-fsc');
  if (!el) return;
  const printers = getFlt();
  const n = printers.length || 1;
  const cols = calcFSCols(n, 16/9);
  const rows = Math.ceil(n / cols);

  clearGridDOM();
  adoptForFullscreen(printers);

  el.className = 'view fs-view';
  el.innerHTML = '<button class="fs-esc" onclick="exitFS()">✕ ESC</button>'
    + '<div class="fs-grid-tiles" style="--cols:'+cols+';--cellh:'+Math.floor(100/rows)+'vh">'
    + printers.map(p => buildTile(p)).join('')
    + CAMS.map(c => buildCamTile(c)).join('')
    + '</div>';

  requestAnimationFrame(() => { mountStreams(printers); mountCamStreams(); });
}


function exitFS() {
  ['view-fsv','view-fsc'].forEach(id => {
    const v = document.getElementById(id);
    if (v) v.innerHTML = ''; // IDs freigeben, Streams bleiben in STREAMS bestehen
  });
  setView('grid');
  resumeAllTiles();
}

// resumeAllTiles brings all visible tiles back to running after leaving a
// fullscreen view — previously tiles stayed dark that were
// Vollbild auf "idle" standen.
function resumeAllTiles() {
  const vis = getFlt();
  if (SNAP_MODE > 0) SNAP_AUTO = true;
  setTileStates(vis, activeMode());
  renderGrid();
}

document.addEventListener('keydown', e => {
  if(e.key==='Escape') {
    const fsv=document.getElementById('view-fsv');
    const fsc=document.getElementById('view-fsc');
    if((fsv&&fsv.style.display!=='none')||(fsc&&fsc.style.display!=='none')) exitFS();
  }
});

function setAllResolution(res) {
  fetch('/api/camera/resolution',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({resolution:res})})
    .then(r=>r.json()).then(d=>{
      toast('Kamera '+res+' gesetzt ('+(d.sent||0)+' Drucker)','ok');
      document.getElementById('btn-720p').classList.toggle('res-active',res==='720p');
      document.getElementById('btn-1080p').classList.toggle('res-active',res==='1080p');
    }).catch(()=>{});
}


// ─── FILE SEARCH ACROSS MULTIPLE PRINTERS ─────────────────────────────────────
//
// The search runs as a background operation: hits appear as soon as the
// respective printer has answered, and it can be cancelled. The search runs
// always within the current printer filter — "all printers" is itself a filter.

let SEARCH_JOB = null, SEARCH_POLL = null;
let SEARCH_HITS = {};   // ip -> Map(schluessel -> Treffer)
let SEARCH_SEEN = {hits: 0, failed: 0};
let SEARCH_BUSY = false; // verhindert ueberlappende Abfragen
let SEARCH_SEL = new Set(); // "ip\u0000datei"

function searchKey(ip, file) { return ip + '\u0000' + file; }

function onSearchTyping() {
  const q = (document.getElementById('fileSearch').value || '').trim();
  const btn = document.getElementById('btn-search');
  if (SEARCH_JOB) return; // laufende Suche nicht anfassen
  btn.disabled = q.length < 3;
  setText('fileSearchStatus', (q.length && q.length < 3) ? 'noch ' + (3 - q.length) + ' Zeichen' : '');
}

function toggleSearch() {
  if (SEARCH_JOB) { stopSearch(); return; }
  startSearch();
}

async function startSearch() {
  const q = (document.getElementById('fileSearch').value || '').trim();
  if (q.length < 3) return;

  clearSearchResults();
  const targets = syncFilteredPrinters();
  if (!targets.length) { setText('fileSearchStatus', 'Drucker wählen'); toast('Drucker wählen','er'); return; }

  busyStart('durchsucht Drucker …');
  try {
    const r = await fetch('/api/sync/search/start', {
      method: 'POST', headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({q, ips: targets.map(p => p.ip)})
    });
    if (!r.ok) throw new Error((await r.text()).trim() || ('HTTP ' + r.status));
    const d = await r.json();
    SEARCH_JOB = d.id;
    SEARCH_SEEN = {hits: 0, failed: 0};
    setSearchRunning(true, d.total);
    SEARCH_POLL = setInterval(pollSearch, 500);
    pollSearch();
  } catch(e) {
    setText('fileSearchStatus', e.message || String(e));
  } finally { busyEnd(); }
}

async function stopSearch() {
  if (!SEARCH_JOB) return;
  try { await fetch('/api/job/stop?id=' + encodeURIComponent(SEARCH_JOB), {method:'POST'}); } catch(e) {}
  setText('fileSearchStatus', 'wird abgebrochen …');
}

function setSearchRunning(running, total) {
  const btn = document.getElementById('btn-search');
  btn.textContent = running ? 'Suche stoppen' : 'Suchen';
  btn.disabled = false;
  btn.classList.toggle('primary', !running);
  setShown(document.getElementById('searchBarWrap'), running);
  if (running) {
    document.getElementById('searchBar').style.setProperty('--w', '0%');
    setText('fileSearchStatus', '0 von ' + total + ' Druckern');
  }
}

function clearSearchResults() {
  SEARCH_HITS = {};
  SEARCH_BUSY = false;
  SEARCH_SEL.clear();
  document.querySelectorAll('.sd-search').forEach(el => { el.innerHTML = ''; });
  document.querySelectorAll('.sd-hitcount').forEach(el => { el.textContent = ''; el.className = 'sd-hitcount'; });
  updateSelBar();
  setText('searchDelStatus', '');
}

// syncFilteredPrinters returns the checked printers. Earlier versions
// simply took the model filter — so the search ran over all 42
// devices although only three were selected, and took correspondingly long.
function syncSelectedPrinters() {
  const ips = new Set([...document.querySelectorAll('.sync-cb:checked')].map(cb => cb.dataset.ip));
  return P.filter(p => ips.has(p.ip));
}
function syncFilteredPrinters() {
  return syncSelectedPrinters();
}

async function pollSearch() {
  if (!SEARCH_JOB) return;
  // The query sometimes takes longer than the cadence. Without this lock,
  // two queries started with the same counter and the same hits landed
  // multiple times in the list — that caused 48 entries instead of 3.
  if (SEARCH_BUSY) return;
  SEARCH_BUSY = true;
  let d;
  try {
    const r = await fetch('/api/job/status?id=' + encodeURIComponent(SEARCH_JOB)
      + '&hits=' + SEARCH_SEEN.hits + '&failed=' + SEARCH_SEEN.failed, {cache:'no-store'});
    if (!r.ok) throw new Error('HTTP ' + r.status);
    d = await r.json();
  } catch(e) {
    return;
  } finally {
    SEARCH_BUSY = false;
  }

  SEARCH_SEEN.hits = d.total_hits;
  SEARCH_SEEN.failed = d.total_failed;

  // Storage by key: the same hit cannot appear twice at all,
  // even if the server returns it twice for some reason.
  (d.hits || []).forEach(h => {
    const m = SEARCH_HITS[h.ip] || (SEARCH_HITS[h.ip] = new Map());
    m.set(searchKey(h.ip, h.file), h);
  });
  const touched = new Set((d.hits || []).map(h => h.ip));
  touched.forEach(renderPrinterHits);
  (d.failed || []).forEach(f => markPrinterFailed(f.ip, f.error));

  const pct = d.total ? Math.round(d.done / d.total * 100) : 0;
  document.getElementById('searchBar').style.setProperty('--w', pct + '%');
  document.getElementById('fileSearchStatus').textContent =
    d.done + ' von ' + d.total + ' Druckern · ' + d.total_hits + ' Treffer'
    + (d.total_failed ? ' · ' + d.total_failed + ' nicht erreichbar' : '');

  if (!d.complete) return;
  clearInterval(SEARCH_POLL); SEARCH_POLL = null;
  SEARCH_JOB = null;
  setSearchRunning(false);
  document.getElementById('fileSearchStatus').textContent =
    (d.stopped ? 'abgebrochen — ' : '') + d.total_hits + ' Treffer auf ' + d.done + ' Druckern'
    + (d.total_failed ? ', ' + d.total_failed + ' nicht erreichbar' : '');
  // Mark printers without hits
  syncFilteredPrinters().forEach(p => {
    const el = document.getElementById('sdhits-' + p.ip);
    if (el && !el.textContent) { el.textContent = 'keine Treffer'; el.className = 'sd-hitcount none'; }
  });
}

function hitsOf(ip) {
  const m = SEARCH_HITS[ip];
  return m ? [...m.values()] : [];
}

function renderPrinterHits(ip) {
  const box = document.getElementById('sdsearch-' + ip);
  const count = document.getElementById('sdhits-' + ip);
  const hits = hitsOf(ip);
  if (count) { count.textContent = hits.length + ' Treffer'; count.className = 'sd-hitcount'; }
  if (!box) return;
  box.innerHTML = hits.map(h => {
    const k = searchKey(h.ip, h.file);
    return '<label class="fs-hit">'
      + '<input type="checkbox" ' + (SEARCH_SEL.has(k) ? 'checked ' : '')
      + 'onchange="toggleHit(this,\'' + esc(h.ip) + '\',\'' + esc(h.file).replace(/'/g, "&#39;") + '\')">'
      + '<span class="fs-file" title="' + esc(h.file) + '">' + esc(h.file) + '</span>'
      + '<span class="fs-size">' + fmtSize(h.size) + '</span>'
      + '</label>';
  }).join('');
}

function markPrinterFailed(ip, reason) {
  const count = document.getElementById('sdhits-' + ip);
  if (count) { count.textContent = 'nicht erreichbar'; count.className = 'sd-hitcount fail'; count.title = reason || ''; }
}


// ─── Selection and delete ─────────────────────────────────────────────────────

function toggleHit(el, ip, file) {
  const k = searchKey(ip, file);
  if (el.checked) SEARCH_SEL.add(k); else SEARCH_SEL.delete(k);
  updateSelBar();
}

function selectAllHits(on) {
  SEARCH_SEL.clear();
  if (on) Object.keys(SEARCH_HITS).forEach(ip => hitsOf(ip).forEach(h => SEARCH_SEL.add(searchKey(h.ip, h.file))));
  Object.keys(SEARCH_HITS).forEach(renderPrinterHits);
  updateSelBar();
}

function updateSelBar() {
  const bar = document.getElementById('searchSelBar');
  const any = Object.keys(SEARCH_HITS).length > 0;
  setShown(bar, any);
  const printers = new Set([...SEARCH_SEL].map(k => k.split('\u0000')[0]));
  setText('searchSelCount', SEARCH_SEL.size ? (SEARCH_SEL.size + ' Datei(en) auf ' + printers.size + ' Drucker(n) ausgewählt') : 'nichts ausgewählt');
}

let DEL_JOB = null, DEL_POLL = null;

async function deleteSelectedHits() {
  if (!SEARCH_SEL.size) { toast('Nichts ausgewählt', 'er'); return; }
  const items = [...SEARCH_SEL].map(k => { const [ip, file] = k.split('\u0000'); return {ip, file}; });
  const printers = new Set(items.map(i => i.ip));

  if (!confirm(items.length + ' Datei(en) auf ' + printers.size + ' Drucker(n) löschen?\n\n'
      + 'Das entfernt sie unwiderruflich von den SD-Karten.')) return;

  const status = document.getElementById('searchDelStatus');
  status.textContent = 'löscht …';
  try {
    const r = await fetch('/api/sync/delete/start', {
      method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({items})
    });
    if (!r.ok) throw new Error((await r.text()).trim() || ('HTTP ' + r.status));
    const d = await r.json();
    DEL_JOB = d.id;
    DEL_POLL = setInterval(pollDelete, 600);
  } catch(e) {
    status.textContent = e.message || String(e);
  }
}

async function pollDelete() {
  if (!DEL_JOB) return;
  let d;
  try { d = await (await fetch('/api/job/status?id=' + encodeURIComponent(DEL_JOB) + '&hits=0&failed=0', {cache:'no-store'})).json(); }
  catch(e) { return; }

  const status = document.getElementById('searchDelStatus');
  status.textContent = d.deleted + ' von ' + d.total + ' gelöscht'
    + (d.total_failed ? ', ' + d.total_failed + ' fehlgeschlagen' : '');
  if (!d.complete) return;

  clearInterval(DEL_POLL); DEL_POLL = null; DEL_JOB = null;

  // Remove deleted ones from the hit list; what failed stays
  const stillFailing = new Set((d.failed || []).map(f => searchKey(f.ip, f.file)));
  Object.keys(SEARCH_HITS).forEach(ip => {
    const m = SEARCH_HITS[ip];
    [...m.keys()].forEach(k => {
      if (SEARCH_SEL.has(k) && !stillFailing.has(k)) m.delete(k);
    });
    if (!m.size) delete SEARCH_HITS[ip];
  });
  SEARCH_SEL.clear();
  document.querySelectorAll('.sd-search').forEach(el => { el.innerHTML = ''; });
  Object.keys(SEARCH_HITS).forEach(renderPrinterHits);
  updateSelBar();

  if (d.total_failed) {
    const names = [...new Set((d.failed || []).map(f => f.name || f.ip))].join(', ');
    toast(d.deleted + ' gelöscht, fehlgeschlagen bei: ' + names, 'er');
  } else {
    toast(d.deleted + ' Datei(en) gelöscht', 'ok');
  }
}

// ─── STARTBILD ────────────────────────────────────────────────────────────────
//
// Two seconds of image, then fade out softly. The image is preloaded —
// otherwise a black area flashes on first start because the JPEG is still
// on its way. If it never arrives, the splash disappears anyway;
// it must under no circumstances cover the UI.
(function () {
  const el = document.getElementById('splash');
  if (!el) return;
  const AB = 2000;
  const start = Date.now();
  let weg = false;
  function closeSplashInner() {
    if (weg) return;
    weg = true;
    el.classList.add('weg');
    setTimeout(() => { if (el && el.parentNode) el.parentNode.removeChild(el); }, 500);
  }
  const scharf = document.getElementById('splash-scharf');
  const img = new Image();
  img.onload = () => {
    if (scharf) {
      scharf.style.backgroundImage = 'url(/splash.jpg?v=__BUILD_ID__)';
      scharf.classList.add('da');
    }
  };
  img.onerror = () => {};   // bleibt bei der Vorschau, kein leerer Bildschirm
  img.src = '/splash.jpg?v=__BUILD_ID__';
  // Emergency brake: even if something hangs, it ends after 4 s.
  setTimeout(closeSplashInner, Math.max(AB - (Date.now() - start), 0));
  setTimeout(closeSplashInner, 4000);
})();

// ─── DRUCK STARTEN ────────────────────────────────────────────────────────────
//
// Select a file from the printer, map a tray, start. The file list comes
// from the device itself — typed names deliberately do not exist here.
async function openPrintStart(ip) {
  const box = byId('ds-' + ip);
  if (!box) return;
  if (!box.classList.contains('hidden')) { box.classList.add('hidden'); detailHalten(ip, false); return; }
  box.classList.remove('hidden');
  detailHalten(ip, true);   // ab jetzt nichts mehr unter den Fingern wegziehen
  box.innerHTML = '<div class="note"><span class="spinner spinner-sm"></span> lese Dateien vom Drucker …</div>';
  try {
    const d = await (await fetch('/api/sync/sdlist?ip=' + encodeURIComponent(ip))).json();
    if (d.error) throw new Error(d.error);
    const dateien = (d.files || []).filter(f => /\.(3mf|gcode)$/i.test(f.name));
    if (!dateien.length) { box.innerHTML = '<div class="note">Keine druckbaren Dateien auf dem Gerät.</div>'; return; }
    drawPrintStart(ip, dateien);
  } catch(e) {
    box.innerHTML = '<div class="dd-err fs12">' + esc(e.message || String(e)) + '</div>';
  }
}

function drawPrintStart(ip, dateien) {
  const box = byId('ds-' + ip);
  const p = P.find(x => x.ip === ip) || {};
  const s = STATUS[ip] || {};
  // Offer all AMS units, including empty trays — with two AMS one should
  // see both and also be able to pick a still-empty tray. The value is
  // 'ams:tray' so the mapping stays unambiguous.
  const faecher = [];
  (s.ams || []).forEach((u, ui) => {
    for (let ti = 0; ti < 4; ti++) {
      const tr = (u.trays || [])[ti] || {};
      const wo = 'AMS ' + (ui + 1) + ' · Fach ' + (ti + 1);
      faecher.push({wert: ui + ':' + ti, text: wo + (tr.type ? ' · ' + tr.type : ' · leer')});
    }
  });

  box.innerHTML =
      '<div class="ds-zeile"><label class="ds-lbl">Datei</label>'
    + '<select class="ds-feld" id="ds-datei-' + esc(ip) + '">'
    + dateien.map(f => '<option value="' + esc(f.name) + '">' + esc(f.name) + '</option>').join('')
    + '</select></div>'
    + (faecher.length
        ? '<div class="ds-map-titel">Filamentzuordnung</div>'
          + [1,2,3,4].map(n =>
              '<div class="ds-zeile"><label class="ds-lbl">Farbe ' + n + '</label>'
              + '<select class="ds-feld ds-map" id="ds-map' + n + '-' + esc(ip) + '">'
              + '<option value="-1">' + (n === 1 ? 'externe Rolle / automatisch' : 'nicht zuordnen') + '</option>'
              + faecher.map(f => '<option value="' + esc(f.wert) + '">' + esc(f.text) + '</option>').join('')
              + '</select></div>').join('')
          + '<div class="ds-hinweis">Einfarbig: nur Farbe 1 setzen. Mehrfarbig: den Farben der Datei der Reihe nach Fächer zuordnen.</div>'
        : '<div class="ds-hinweis">Kein AMS gemeldet — es wird die externe Rolle verwendet.</div>')
    + '<div class="ds-zeile"><label class="ds-check"><input type="checkbox" id="ds-level-' + esc(ip) + '" checked> Kalibrieren</label>'
    + '<label class="ds-check"><input type="checkbox" id="ds-tl-' + esc(ip) + '"> Zeitraffer</label></div>'
    + '<div class="ds-zeile"><button class="pc-btn" data-ip="' + esc(ip) + '" onclick="startPrint(this.dataset.ip)">▶ Druck starten</button>'
    + '<span class="note-sm" id="ds-status-' + esc(ip) + '"></span></div>';
}

async function startPrint(ip) {
  const p = P.find(x => x.ip === ip) || {name: ip};
  const datei = (byId('ds-datei-' + ip) || {}).value;
  if (!datei) return;
  // Mapping from the up to four color fields. Value '-1' = do not map,
  // otherwise 'ams:tray' -> global index ams*4 + tray.
  const mapping = [1,2,3,4].map(n => {
    const el = byId('ds-map' + n + '-' + ip);
    if (!el || el.value === '-1') return -1;
    const [a, t] = el.value.split(':').map(x => parseInt(x, 10));
    return a * 4 + t;
  });
  // Trim trailing "not mapped" so the list is as long as
  // the number of used colors.
  while (mapping.length && mapping[mapping.length - 1] < 0) mapping.pop();
  const genutzt = mapping.some(v => v >= 0);
  const zusammenfassung = genutzt
    ? mapping.map((v, i) => 'Farbe ' + (i+1) + ' → ' + (v < 0 ? '—' : 'Fach ' + (v+1))).join('\n')
    : 'externe Rolle / automatisch';
  if (!confirm('Auf "' + p.name + '" drucken?\n\n' + datei + '\n\n' + zusammenfassung)) return;

  const stat = byId('ds-status-' + ip);
  if (stat) stat.innerHTML = '<span class="spinner spinner-sm"></span> sendet …';
  busyStart('startet Druck …');
  try {
    const r = await fetch('/api/print/start', {
      method: 'POST', headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({ip, datei, mapping: mapping,
                            use_ams: mapping.some(v => v >= 0),
                            leveling: !!(byId('ds-level-' + ip) || {}).checked,
                            timelapse: !!(byId('ds-tl-' + ip) || {}).checked,
                            flow_cali: true})
    });
    const d = await r.json().catch(() => ({}));
    if (!r.ok) throw new Error(d.error || (await r.text()) || ('HTTP ' + r.status));
    if (d.ok === false) throw new Error(d.error || 'abgelehnt');
    if (stat) stat.textContent = '';
    toast(p.name + ': Druck gestartet ✓', 'ok');
    byId('ds-' + ip).classList.add('hidden');
    detailHalten(ip, false);
    setTimeout(loadStatus, 1500);
  } catch(e) {
    if (stat) stat.textContent = '';
    toast('Start fehlgeschlagen: ' + (e.message || e), 'er');
  } finally { busyEnd(); }
}

// ─── ABOUT DRUCKERFARM ────────────────────────────────────────────────────────
//
// The splash again, on request: centered, with the copyright from the
// settings. The image is already loaded, so it appears without delay.
function showSplash() {
  const el = byId('ueber');
  if (!el) return;
  el.style.backgroundImage = 'url(/splash.jpg?v=__BUILD_ID__)';
  // Take the copyright exactly as it stands in the settings — a
  // zweite Fassung im Code waere sofort veraltet.
  const quelle = document.querySelector('.copyright-card p');
  setText('ueber-copy', quelle ? quelle.textContent.trim() : '');
  const ver = byId('upd-current');
  setText('ueber-version', ver && ver.textContent ? 'Version ' + ver.textContent.trim() : '');
  el.classList.remove('hidden');
}

function closeSplash() {
  const el = byId('ueber');
  if (el) el.classList.add('hidden');
}
document.addEventListener('keydown', e => { if (e.key === 'Escape') closeSplash(); });

// ─── UPLOAD-HISTORIE ──────────────────────────────────────────────────────────
//
// Stored permanently so that even after days it is traceable what
// went where — and what went wrong.
async function loadUploadLog() {
  const box = byId('ul-liste');
  if (!box) return;
  box.innerHTML = '<div class="note"><span class="spinner spinner-sm"></span> lädt …</div>';
  try {
    const d = await (await fetch('/api/upload-log', {cache:'no-store'})).json();
    drawUploadLog(d);
  } catch(e) {
    box.innerHTML = '<div class="dd-err fs12">' + esc(e.message || String(e)) + '</div>';
  }
}

function drawUploadLog(d) {
  const box = byId('ul-liste');
  const eintraege = d.eintraege || [];
  setText('ul-zusammenfassung', eintraege.length
    ? eintraege.length + ' Einträge · ' + d.erfolge + ' erfolgreich · ' + d.fehler + ' fehlgeschlagen'
    : 'noch nichts hochgeladen');
  setDis('btn-ul-leeren', eintraege.length === 0);
  if (!eintraege.length) { box.innerHTML = '<div class="note">Die Historie ist leer.</div>'; return; }
  box.innerHTML = eintraege.map(e => {
    const dt = new Date(e.zeit);
    const zeit = isNaN(dt) ? '' : dt.toLocaleString('de-CH',
      {day:'2-digit',month:'2-digit',year:'2-digit',hour:'2-digit',minute:'2-digit'});
    return '<div class="ul-zeile' + (e.erfolg ? '' : ' fehlgeschlagen') + '">'
      + '<span class="ul-zeit">' + esc(zeit) + '</span>'
      + '<span class="ul-drucker">' + esc(e.name || e.ip) + '</span>'
      + '<span class="ul-datei" title="' + esc(e.datei) + '">' + esc(e.datei) + '</span>'
      + '<span class="ul-groesse">' + (e.bytes ? fmtSize(e.bytes) : '') + '</span>'
      + '<span class="ul-ergebnis">' + (e.erfolg ? '✓' : '✗ ' + esc(e.fehler || 'Fehler')) + '</span>'
      + '</div>';
  }).join('');
}

async function leereUploadLog() {
  if (!confirm('Die gesamte Upload-Historie löschen?\n\nDie hochgeladenen Dateien bleiben auf den Druckern — es verschwindet nur die Aufstellung.')) return;
  try {
    const d = await (await fetch('/api/upload-log', {method:'DELETE'})).json();
    toast((d.geleert || 0) + ' Einträge gelöscht', 'ok');
    loadUploadLog();
  } catch(e) { toast(e.message || String(e), 'er'); }
}

// ─── ERREICHBARKEIT ───────────────────────────────────────────────────────────
//
// Answers the question one can otherwise only guess: is it the device, the
// camera or the program? The three ports are checked individually.
async function checkReach() {
  const btn = byId('btn-reach'), box = byId('reach-box'), list = byId('reach-list');
  if (!btn || !box) return;
  setShown(box, true);
  btn.disabled = true;
  btn.innerHTML = '<span class="spinner spinner-sm"></span> prüft …';
  list.innerHTML = '<div class="note"><span class="spinner spinner-sm"></span> klopft an allen Druckern an …</div>';
  setText('reach-sum', '');
  busyStart('prüft Erreichbarkeit …');
  try {
    const d = await (await fetch('/api/reach', {cache: 'no-store'})).json();
    renderReach(d);
  } catch (e) {
    list.innerHTML = '<div class="dd-err fs12">' + esc(e.message || String(e)) + '</div>';
  } finally {
    busyEnd();
    btn.disabled = false;
    btn.textContent = '📡 Drucker erreichbar?';
  }
}

function reachKlasse(u) {
  if (u.indexOf('erreichbar — Bild') >= 0) return 'gut';
  if (u.indexOf('aus oder nicht im Netz') >= 0) return 'schlecht';
  return 'warn';
}

function renderReach(d) {
  const list = byId('reach-list');
  const rows = d.drucker || [];
  if (!rows.length) { list.innerHTML = '<div class="note">Keine Drucker angelegt.</div>'; return; }
  list.innerHTML = rows.map(r =>
      '<div class="rc-row">'
    + '<span>' + esc(r.name || r.ip) + '</span>'
    + '<span class="rc-ports">' + (r.ports || []).map(p =>
        '<span class="rc-p ' + (p.offen ? 'auf' : 'zu') + '" title="'
        + esc(p.was + ' · Port ' + p.port + (p.offen ? ' offen (' + p.ms + ' ms)' : ' ' + (p.grund || ''))) + '">'
        + esc(p.was) + '</span>').join('')
    + '</span>'
    + '<span class="rc-urteil ' + reachKlasse(r.urteil || '') + '">' + esc(r.urteil || '') + '</span>'
    + '</div>').join('');
  const zus = d.zusammenfassung || {};
  setText('reach-sum', Object.keys(zus).sort((a, b) => zus[b] - zus[a])
    .map(k => zus[k] + '× ' + k).join('   ·   '));
}

// ─── PROGRAMMVERSION ──────────────────────────────────────────────────────────
let UPDATE_INFO = null, UPD_POLL = null;

async function openDataFolder() {
  try {
    const r = await fetch('/api/open-folder', {method:'POST', headers:{'Content-Type':'application/json'}, body:'{}'});
    if (!r.ok) throw new Error((await r.text()).trim());
    toast('Ordner geöffnet', 'ok');
  } catch(e) { toast(e.message || String(e), 'er'); }
}

async function loadUpdateConfig() {
  try {
    const d = await (await fetch('/api/update/config', {cache:'no-store'})).json();
    const set = (id, v) => { const el = document.getElementById(id); if (el && el !== document.activeElement) el.value = v || ''; };
    set('upd-repo', d.repo);
    const cur = document.getElementById('upd-current');
    if (cur) cur.textContent = d.version + (d.version === 'dev' ? '  (ohne Versionsstempel gebaut)' : '') + ' · Build ' + BUILD_ID;
    const dir = document.getElementById('upd-datadir');
    if (dir) dir.textContent = d.data_dir || '–';
    const yp = document.getElementById('yaml-path');
    if (yp && d.data_dir) {
      yp.textContent = d.data_dir + '\\go2rtc.yaml';
      yp.title = 'Ordner im Explorer öffnen';
    }
    return d;
  } catch(e) { return null; }
}

async function saveUpdateRepo() {
  const repo = (document.getElementById('upd-repo').value || '').trim();
  const body = {repo, token: ''};
  try {
    const r = await fetch('/api/update/config', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify(body)});
    if (!r.ok) throw new Error((await r.text()).trim());
    toast('Repository gespeichert', 'ok');
    checkUpdate(true);
  } catch(e) { toast(e.message || String(e), 'er'); }
}

async function checkUpdate(loud) {
  const status = document.getElementById('upd-status');
  const btn = document.getElementById('btn-updinstall');
  await loadUpdateConfig();
  if (status) status.textContent = loud ? 'suche …' : '';
  try {
    const d = await (await fetch('/api/update/check', {cache:'no-store'})).json();
    UPDATE_INFO = d;
    const notes = document.getElementById('upd-notes');
    if (d.error) {
      if (status) status.textContent = loud ? d.error : '';
      if (btn) setShown(btn, false);
      if (notes) notes.textContent = '';
      return;
    }
    if (d.available) {
      if (status) status.textContent = 'Version ' + d.latest + ' verfügbar';
      if (btn) setShown(btn, true);
      if (notes) notes.textContent = (d.notes || '').substring(0, 700);
      toast('Update verfügbar: ' + d.latest, 'ok');
    } else {
      if (status) status.textContent = loud ? 'Es läuft die neueste Fassung' : '';
      if (btn) setShown(btn, false);
      if (notes) notes.textContent = '';
    }
  } catch(e) {
    if (status && loud) status.textContent = e.message || String(e);
  }
}

async function installUpdate() {
  if (!UPDATE_INFO || !UPDATE_INFO.available) return;
  if (!confirm('Auf Version ' + UPDATE_INFO.latest + ' aktualisieren?\n\nDas Programm lädt die neue Fassung, ersetzt sich selbst und startet neu. Die Vorgängerversion bleibt als Rückweg liegen.')) return;

  const btn = document.getElementById('btn-updinstall');
  btn.disabled = true;
  setShown(document.getElementById('upd-progress'), true);
  setText('upd-status', '');
  try {
    const r = await fetch('/api/update/install', {method:'POST'});
    if (!r.ok) throw new Error((await r.text()).trim() || ('HTTP ' + r.status));
  } catch(e) {
    btn.disabled = false;
    setText('upd-status', e.message || String(e));
    return;
  }
  if (UPD_POLL) clearInterval(UPD_POLL);
  UPD_POLL = setInterval(pollUpdateProgress, 600);
}

const UPD_STEPS = {download:'Lade', verify:'Prüfe', replace:'Ersetze', done:'Fertig', error:'Fehler'};

async function pollUpdateProgress() {
  let p;
  try { p = await (await fetch('/api/update/progress', {cache:'no-store'})).json(); } catch(e) { return; }
  const bar = document.getElementById('upd-bar');
  const txt = document.getElementById('upd-progress-text');
  const label = UPD_STEPS[p.step] || p.step;

  if (p.total > 0) {
    const pct = Math.min(100, Math.round(p.received / p.total * 100));
    bar.style.setProperty('--w', pct + '%');
    txt.textContent = label + ' — ' + fmtMB(p.received) + ' / ' + fmtMB(p.total) + ' (' + pct + '%)';
  } else {
    bar.style.setProperty('--w', p.step === 'download' ? '50%' : '85%');
    txt.textContent = label + (p.received ? ' — ' + fmtMB(p.received) : '') + '…';
  }
  if (!p.done) return;

  clearInterval(UPD_POLL); UPD_POLL = null;
  if (p.error) {
    setDis('btn-updinstall', false);
    setText('upd-status', p.error);
    return;
  }
  bar.style.setProperty('--w', '100%');
  txt.textContent = 'Fertig — die neue Fassung startet. Dieses Fenster kann geschlossen werden.';
}

// ─── NETZWERKSUCHE ────────────────────────────────────────────────────────────
let SCAN_RUNNING = false;

async function scanNetwork() {
  if (SCAN_RUNNING) return;
  SCAN_RUNNING = true;
  const btn = document.getElementById('btn-scan');
  const status = document.getElementById('scan-status');
  const box = document.getElementById('scan-results');
  btn.disabled = true;
  btn.textContent = '📡 Suche läuft …';
  status.textContent = 'sendet Suchpakete, wartet auf Antworten …';
  box.innerHTML = '';

  try {
    const r = await fetch('/api/discover?timeout=5', {cache: 'no-store'});
    if (!r.ok) throw new Error((await r.text()).trim() || ('HTTP ' + r.status));
    const d = await r.json();
    renderScanResults(d.printers || []);
    const e = d.empfang || {};
    const teile = [(d.printers || []).length + ' Gerät(e)'];
    if (d.aus_mithoeren) teile.push(d.aus_mithoeren + ' nur über Mithören');
    if (d.umgezogen) teile.push(d.umgezogen + ' mit neuer Adresse');
    status.innerHTML = esc(teile.join(' · '))
      + '<br><span class="dd-fein">Empfang: ' + esc((e.empfangswege || []).join(', ') || 'keiner')
      + ' · ' + (e.pakete_gesamt || 0) + ' Pakete gehört'
      + ((e.nicht_moeglich || []).length ? ' · nicht möglich: ' + esc(e.nicht_moeglich.join(', ')) : '')
      + '</span>';
  } catch(e) {
    status.textContent = '';
    box.innerHTML = '<div class="dd-err fs12">' + esc(e.message || String(e)) + '</div>';
  } finally {
    SCAN_RUNNING = false;
    btn.disabled = false;
    btn.textContent = '📡 Netzwerk nach Druckern durchsuchen';
  }
}

// SCAN_FOUND holds the last search yield. Adoption is from this state
// and not from the DOM — otherwise adoption depends on the presentation.
let SCAN_FOUND = [];

function renderScanResults(list) {
  const box = document.getElementById('scan-results');
  SCAN_FOUND = list || [];
  if (!list.length) {
    box.innerHTML = '<div class="note lh">'
      + 'Nichts gefunden. Möglich ist, dass der Rechner in einem anderen Netz hängt oder '
      + 'die Switches keine Multicast-Pakete weiterreichen. Dann bleibt der Weg über CSV oder manuelle Eingabe.'
      + '</div>';
    return;
  }

  // Known is whoever's serial number is already in the list. The address
  // is unsuitable for that — it is exactly what changes.
  const nachSerie = {};
  P.forEach(p => { if (p.serial) nachSerie[p.serial.trim()] = p; });

  const neue = [], bekannte = [];
  list.forEach(d => {
    const p = d.serial ? nachSerie[d.serial.trim()] : null;
    (p ? bekannte : neue).push({d, p});
  });

  const sortiere = arr => arr.sort((a, b) =>
    (a.p ? a.p.name : a.d.name || a.d.ip).localeCompare(b.p ? b.p.name : b.d.name || b.d.ip));
  sortiere(neue); sortiere(bekannte);

  // Already-added devices are only shown if they moved
  // — otherwise the list is a wall of the known in which the new gets lost.
  const umgezogene = bekannte.filter(x => x.p.ip !== x.d.ip);
  // Known, same address, but the printer reports a different name than
  // the one stored in the program — e.g. because it was renamed in the
  // vendor app. Show these cases so the name can be adopted.
  const namensAenderung = bekannte.filter(x => x.p.ip === x.d.ip
    && (x.d.name || '').trim() && (x.d.name || '').trim() !== (x.p.name || '').trim());

  box.innerHTML =
      (neue.length
        ? '<div class="scan-bar">'
          + '<label class="scan-check"><input type="checkbox" id="scan-all" onchange="scanToggleAll(this.checked)"> alle ' + neue.length + ' auswählen</label>'
          + '<span class="note-sm" id="scan-sel-count">0 ' + t('selected') + '</span>'
          + '<label class="scan-check"><input type="checkbox" id="scan-adopt-name" checked> Gerätenamen übernehmen</label>'
          + '<button class="hbtn primary push-right" id="btn-adopt-sel" onclick="adoptSelected()" disabled>Ausgewählte hinzufügen</button>'
          + '</div>'
        : '')
    + (neue.length ? abschnitt('Noch nicht in der Liste', neue.map(x => rowNew(x.d)).join('')) : '')
    + (umgezogene.length ? abschnitt('Neue Adresse', umgezogene.map(x => rowKnown(x.d, x.p)).join('')) : '')
    + (namensAenderung.length ? abschnitt('Name im Drucker geändert', namensAenderung.map(x => rowNameChanged(x.d, x.p)).join('')) : '')
    + ((!neue.length && !umgezogene.length && !namensAenderung.length)
        ? '<div class="note">' + list.length + ' Gerät(e) gefunden, alle bereits angelegt und unter bekannter Adresse.</div>' : '');
  updateScanSel();
}

function abschnitt(titel, inhalt) {
  return '<div class="scan-abschnitt"><div class="scan-abschnitt-titel">' + esc(titel) + '</div>' + inhalt + '</div>';
}

// namePaar shows the name from the program and behind it in parentheses the one
// the device itself reports. If none is stored, the device's applies.
function namePaar(programmName, geraetName) {
  const pn = (programmName || '').trim(), gn = (geraetName || '').trim();
  if (!pn) return esc(gn || '(ohne Namen)');
  if (!gn || gn === pn) return esc(pn);
  return esc(pn) + ' <span class="dd-fein">(' + esc(gn) + ')</span>';
}

function rowNew(d) {
  return '<div class="scan-row">'
    + '<input type="checkbox" class="scan-cb" data-ip="' + esc(d.ip) + '" onchange="updateScanSel()">'
    + '<span class="scan-model">' + esc(d.model || '?') + '</span>'
    + '<span>' + namePaar('', d.name) + '</span>'
    + '<span class="scan-mono">' + esc(d.ip) + '</span>'
    + '<span class="scan-mono" title="Seriennummer">' + esc(d.serial || '–') + '</span>'
    + '<span class="scan-tag ' + esc(d.bind || '') + '">'
      + esc(d.bind === 'free' ? 'frei' : (d.bind === 'occupied' ? 'belegt' : (d.bind || '–'))) + '</span>'
    + '<input class="scan-code" placeholder="Zugangscode (später möglich)" data-ip="' + esc(d.ip) + '" oninput="updateScanSel()">'
    + '</div>';
}

function rowKnown(d, p) {
  const umgezogen = p.ip !== d.ip;
  return '<div class="scan-row bekannt' + (umgezogen ? ' umgezogen' : '') + '">'
    + '<span class="scan-haken">' + (umgezogen ? '↻' : '✓') + '</span>'
    + '<span class="scan-model">' + esc(p.model || d.model || '?') + '</span>'
    + '<span>' + namePaar(p.name, d.name) + '</span>'
    + '<span class="scan-mono">' + esc(d.ip)
      + (umgezogen ? '<br><span class="dd-fein">eingetragen: ' + esc(p.ip) + '</span>' : '') + '</span>'
    + '<span class="scan-mono" title="Seriennummer">' + esc(d.serial || '–') + '</span>'
    + '<span class="scan-tag ' + esc(d.bind || '') + '">'
      + esc(d.bind === 'free' ? 'frei' : (d.bind === 'occupied' ? 'belegt' : (d.bind || '–'))) + '</span>'
    + (umgezogen
        ? '<button class="hbtn primary" data-alt="' + esc(p.ip) + '" data-neu="' + esc(d.ip)
          + '" onclick="applyNewIP(this.dataset.alt, this.dataset.neu)">IP übernehmen</button>'
        : '<span class="note-sm">unverändert</span>')
    + '</div>';
}

// rowNameChanged shows a known printer whose device name differs from the
// stored one — with a button to adopt the device name into the program.
function rowNameChanged(d, p) {
  return '<div class="scan-row bekannt">'
    + '<span class="scan-haken">✎</span>'
    + '<span class="scan-model">' + esc(p.model || d.model || '?') + '</span>'
    + '<span>' + namePaar(p.name, d.name) + '</span>'
    + '<span class="scan-mono">' + esc(d.ip) + '</span>'
    + '<span class="scan-mono" title="Seriennummer">' + esc(d.serial || '–') + '</span>'
    + '<button class="hbtn primary" data-ip="' + esc(p.ip) + '" data-name="' + esc(d.name)
      + '" onclick="applyDeviceName(this.dataset.ip, this.dataset.name)">Namen übernehmen</button>'
    + '</div>';
}

// applyDeviceName adopts the name reported by the printer into the
// gespeicherte Druckerliste.
async function applyDeviceName(ip, neuerName) {
  const p = P.find(x => x.ip === ip);
  if (!p || !neuerName) return;
  const updated = {model: p.model, name: neuerName.trim(), ip: p.ip, code: p.code, serial: p.serial, fav: p.fav, added: p.added};
  try {
    const r = await fetch('/api/printers/' + encodeURIComponent(ip), {
      method: 'PUT', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(updated)
    });
    if (!r.ok) throw new Error((await r.text()).trim() || ('HTTP ' + r.status));
    Object.assign(p, updated);
    render();
    if (typeof renderSyncPrinterList === 'function') renderSyncPrinterList();
    toast(neuerName.trim() + ': Name übernommen', 'ok');
    if (SCAN_FOUND) renderScanResults(SCAN_FOUND);
  } catch(e) { toast(e.message || String(e), 'er'); }
}

function scanToggleAll(on) {
  document.querySelectorAll('.scan-cb').forEach(cb => { cb.checked = !!on; });
  updateScanSel();
}

// The access code is now optional: a printer may be added without one and
// the code added later. Without a code there is no image and no
// file access — status and control do not need it.
function updateScanSel() {
  const sel = [...document.querySelectorAll('.scan-cb:checked')];
  setText('scan-sel-count', sel.length + ' ' + t('selected'));
  const ohneCode = sel.filter(cb => {
    const inp = document.querySelector('.scan-code[data-ip="' + CSS.escape(cb.dataset.ip) + '"]');
    return !inp || !inp.value.trim();
  }).length;
  const btn = document.getElementById('btn-adopt-sel');
  if (btn) {
    btn.disabled = sel.length === 0;
    btn.textContent = sel.length
      ? (ohneCode ? sel.length + ' hinzufügen (' + ohneCode + '× ohne Code)' : sel.length + ' hinzufügen')
      : 'Ausgewählte hinzufügen';
  }
}

// adoptSelected creates the checked devices — without confirmation, because one
// usually does this for several in a row and each entry afterwards
// bearbeitbar bleibt.
async function adoptSelected() {
  const sel = [...document.querySelectorAll('.scan-cb:checked')].map(cb => cb.dataset.ip);
  if (!sel.length) return;
  const adoptName = !!(document.getElementById('scan-adopt-name') || {}).checked;

  const btn = document.getElementById('btn-adopt-sel');
  if (btn) { btn.disabled = true; btn.innerHTML = '<span class="spinner spinner-sm"></span> legt an …'; }
  busyStart('legt Drucker an …');
  let ok = 0, uebersprungen = 0;
  const fehler = [];
  try {
    for (const ip of sel) {
      const d = SCAN_FOUND.find(x => x.ip === ip) || {ip};
      // "Already present" is decided by the serial number, not the IP: an
      // offline printer with a stale IP must not block adding.
      const schon = d.serial ? P.some(p => (p.serial||'').trim() === d.serial.trim()) : P.some(p => p.ip === ip);
      if (schon) { uebersprungen++; continue; }
      const inp = document.querySelector('.scan-code[data-ip="' + CSS.escape(ip) + '"]');
      const netzName = (d.name || '').trim();
      const ersatz = ((d.model || 'Drucker').toUpperCase()) + '-' + ip.split('.').pop();
      try {
        const p = await apiAdd({
          model: (d.model || '').toUpperCase(),
          name: (adoptName && netzName) ? netzName : ersatz,
          ip: ip, code: inp ? inp.value.trim() : '', serial: d.serial || '',
          fav: false, added: Date.now()
        });
        P.push(p); ST[p.ip] = 'idle';
        ok++;
      } catch(e) { fehler.push(ip + ': ' + (e.message || e)); }
    }
  } finally {
    busyEnd();
    if (btn) btn.disabled = false;
  }
  render();
  renderScanResults(SCAN_FOUND);
  let out = ok + ' Drucker angelegt';
  if (uebersprungen) out += ', ' + uebersprungen + ' bereits vorhanden';
  if (fehler.length) out += ', ' + fehler.length + ' fehlgeschlagen';
  toast(out, fehler.length ? 'er' : 'ok');
  if (fehler.length) console.warn('Nicht angelegt:', fehler);
}

async function applyNewIP(alteIP, neueIP) {
  const p = P.find(x => x.ip === alteIP);
  if (!p) { toast('Drucker nicht mehr in der Liste', 'er'); return; }
  // Without a prompt: the action is reversible and one usually does it
  // for a dozen devices in a row.
  busyStart('übernimmt Adresse …');
  try {
    const r = await fetch('/api/printers/' + encodeURIComponent(alteIP), {
      method: 'PUT', headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({...p, ip: neueIP})
    });
    if (!r.ok) throw new Error((await r.text()).trim() || ('HTTP ' + r.status));
    if (ST[alteIP]) { setTileState(alteIP, 'idle'); delete ST[alteIP]; }
    delete STATUS[alteIP];
    await loadPrinters();
    render();
    // The search stays — one works through the list in order.
    renderScanResults(SCAN_FOUND);
    toast(p.name + ' hört jetzt auf ' + neueIP, 'ok');
  } catch(e) {
    toast(e.message || String(e), 'er');
  } finally { busyEnd(); }
}

// adoptFound pre-fills the manual input row. The access code is missing in the
// SSDP response, so the focus jumps there.
function adoptFound(payload) {
  let d;
  try { d = JSON.parse(payload); } catch(e) { return; }
  const set = (id, v) => { const el = document.getElementById(id); if (el) el.value = v || ''; };
  set('mm', (d.model || '').toUpperCase());
  set('mn', d.name);
  set('mi', d.ip);
  set('ms', d.serial);
  const code = document.getElementById('mc');
  if (code) { code.value = ''; code.focus(); code.scrollIntoView({block: 'center', behavior: 'smooth'}); }
  toast((d.name || d.ip) + ' übernommen — jetzt den Zugangscode eintragen', 'ok');
}

// ─── KOMPONENTEN (go2rtc / ffmpeg) ────────────────────────────────────────────
let COMPONENTS = [], COMP_MISSING = [], COMP_POLL = null, COMP_DOWNLOADABLE = false;

// ─── DIAGNOSE ─────────────────────────────────────────────────────────────────
let DIAG_TEXT = '';

async function showDiagnostics() {
  const box = byId('diag-box'), out = byId('diag-text');
  setShown(box, true);
  if (out) out.textContent = 'sammelt …';
  try {
    // Fetch the health snapshot (aggregate) first, then the detailed diagnostics.
    let health = '';
    try {
      const h = await (await fetch('/api/health', {cache:'no-store'})).json();
      health = formatHealth(h) + '\n';
    } catch(_) {}
    const d = await (await fetch('/api/diagnostics', {cache:'no-store'})).json();
    DIAG_TEXT = health + formatDiagnostics(d);
    if (out) out.textContent = DIAG_TEXT;
  } catch(e) {
    if (out) out.textContent = 'Diagnose fehlgeschlagen: ' + (e.message || e);
  }
}

// formatHealth turns the /api/health aggregate into a short header line:
// how many printers are online, MQTT connections, go2rtc, errors, uptime.
function formatHealth(h) {
  const upt = (s) => {
    s = Math.max(0, s|0);
    const d = Math.floor(s/86400), hh = Math.floor(s%86400/3600), mm = Math.floor(s%3600/60);
    if (d) return d+'d '+hh+'h';
    if (hh) return hh+'h '+mm+'m';
    return mm+'m';
  };
  const L = [];
  L.push('── Gesundheit ─────────────────────────');
  L.push('Drucker online:   ' + (h.printers_online||0) + ' / ' + (h.printers_total||0)
        + '   (MQTT verbunden: ' + (h.mqtt_connected||0) + ')');
  L.push('go2rtc:           ' + (h.go2rtc_running ? 'läuft' : 'AUS')
        + '   Prozesse: ' + (h.go2rtc_count||0)
        + (h.go2rtc_pids && h.go2rtc_pids.length ? '   PIDs: ' + h.go2rtc_pids.join(', ') : ''));
  L.push('Fehler-Drucker:   ' + (h.error_count||0));
  if (h.errors && h.errors.length) {
    h.errors.slice(0, 8).forEach(e => L.push('  • ' + (e.name || e.ip) + ': ' + (e.messages||[]).join('; ')));
    if (h.errors.length > 8) L.push('  … +' + (h.errors.length-8) + ' weitere');
  }
  L.push('Fenster offen:    ' + (h.ui_connected ? 'ja' : 'nein'));
  L.push('Laufzeit Tool:    ' + upt(h.uptime_seconds) + '   Build ' + (h.build||'?'));
  L.push('');
  return L.join('\n');
}

function formatDiagnostics(d) {
  const L = [];
  const yn = v => v ? 'ja' : 'NEIN';
  L.push('Druckerfarm ' + d.version);
  L.push('Datenordner:      ' + d.data_dir);
  L.push('');
  L.push('go2rtc vorhanden: ' + yn(d.go2rtc_present) + '   ' + (d.go2rtc_version || ''));
  L.push('  Datei:          ' + d.go2rtc_binary);
  L.push('  soll laufen:    ' + yn(d.should_run) + '   PID: ' + (d.pid || '–') + '   Neustarts: ' + d.restarts);
  L.push('  Port ' + d.port + ':      ' + (d.port_reachable ? 'erreichbar' : 'NICHT ERREICHBAR'));
  L.push('ffmpeg vorhanden: ' + yn(d.ffmpeg_present) + '   ' + (d.ffmpeg_version || ''));
  L.push('');
  L.push('Drucker eingetragen: ' + d.printers);
  L.push('Streams in go2rtc:   ' + d.streams_loaded + (d.streams_error ? '  (' + d.streams_error + ')' : ''));
  if (d.printers > 0 && d.streams_loaded === 0) {
    L.push('  >>> go2rtc kennt KEINEN Stream — go2rtc.yaml wird nicht angenommen');
  } else if (d.streams_loaded < d.printers) {
    L.push('  >>> es fehlen ' + (d.printers - d.streams_loaded) + ' Streams');
  }
  if ((d.name_collisions || []).length) {
    L.push('  >>> doppelte Streamnamen: ' + d.name_collisions.join(', '));
  }
  L.push('');
  L.push('--- go2rtc.log (letzte Zeilen) ---');
  (d.go2rtc_log || []).forEach(l => L.push(l));
  L.push('');
  L.push('--- druckerfarm.log (letzte Zeilen) ---');
  (d.app_log || []).forEach(l => L.push(l));
  return L.join('\n');
}

async function copyDiagnostics() {
  const hint = byId('diag-hint');
  try {
    await navigator.clipboard.writeText(DIAG_TEXT);
    if (hint) hint.textContent = 'kopiert';
  } catch(e) {
    if (hint) hint.textContent = 'Kopieren nicht möglich — Text markieren und Strg+C';
  }
}

// ─── KOMPONENTEN AKTUALISIEREN ────────────────────────────────────────────────
let COMP_UPDATES = [];

async function checkComponentUpdates(loud) {
  const status = document.getElementById('comp-upd-status');
  const btn = document.getElementById('btn-compupd');
  const runBtn = document.getElementById('btn-compupd-run');
  const box = document.getElementById('comp-updates');
  if (btn) { btn.disabled = true; }
  if (status) status.textContent = loud ? 'prüft …' : '';
  try {
    const d = await (await fetch('/api/components/updates', {cache:'no-store'})).json();
    COMP_UPDATES = d.updates || [];
    const avail = COMP_UPDATES.filter(u => u.available);
    if (box) {
      box.innerHTML = COMP_UPDATES.filter(u => u.available || u.note).map(u =>
        '<div class="cu-row">'
        + '<span class="minw70 cu-name">' + esc(u.label) + '</span>'
        + (u.available
            ? '<span>installiert ' + esc(u.installed || '?') + ' · <span class="cu-new">neu ' + esc(u.latest || '?') + '</span></span>'
            : '<span class="muted cu-note">' + esc(u.note || '') + '</span>')
        + '</div>').join('');
    }
    if (runBtn) setShown(runBtn, avail.length);
    if (status) {
      status.textContent = avail.length
        ? avail.length + ' Aktualisierung(en) verfügbar'
        : (loud ? 'alles aktuell' : '');
    }
  } catch(e) {
    if (status) status.textContent = e.message || String(e);
  } finally {
    if (btn) btn.disabled = false;
  }
}

async function runComponentUpdate() {
  const keys = COMP_UPDATES.filter(u => u.available).map(u => u.key);
  if (!keys.length) return;
  if (!confirm('Aktualisieren: ' + keys.join(', ') + '?\n\n'
      + 'Die neue Fassung wird geladen, geprüft und ersetzt die bisherige. go2rtc startet dabei neu, '
      + 'laufende Videostreams brechen kurz ab.')) return;

  const status = document.getElementById('comp-upd-status');
  const runBtn = document.getElementById('btn-compupd-run');
  runBtn.disabled = true;
  status.textContent = 'lädt …';
  try {
    const r = await fetch('/api/components/updates', {
      method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({components: keys})
    });
    if (!r.ok) throw new Error((await r.text()).trim() || ('HTTP ' + r.status));
  } catch(e) {
    runBtn.disabled = false;
    status.textContent = e.message || String(e);
    return;
  }
  pollComponentUpdate();
}

async function pollComponentUpdate() {
  const status = document.getElementById('comp-upd-status');
  const runBtn = document.getElementById('btn-compupd-run');
  let p;
  try { p = await (await fetch('/api/components/progress', {cache:'no-store'})).json(); }
  catch(e) { setTimeout(pollComponentUpdate, 800); return; }

  const label = SETUP_STEPS[p.step] || p.step;
  status.textContent = p.total > 0
    ? label + ' ' + (p.component || '') + ' — ' + fmtMB(p.received) + ' / ' + fmtMB(p.total)
    : label + ' ' + (p.component || '') + '…';

  if (!p.done) { setTimeout(pollComponentUpdate, 800); return; }
  runBtn.disabled = false;
  if (p.error) { status.textContent = p.error; return; }
  status.textContent = 'aktualisiert';
  setShown(runBtn, false);
  await loadComponents(false);
  await checkComponentUpdates(false);
  setTimeout(chkG2, 1500);
}

let BROWSER_HINT_SHOWN = false;
function showBrowserHint(port) {
  if (BROWSER_HINT_SHOWN) return;
  BROWSER_HINT_SHOWN = true;
  const el = document.getElementById('setup-info');
  const msg = 'Kein Edge oder Chrome gefunden — die Oberfläche läuft im Standardbrowser. '
            + 'Alle Funktionen stehen zur Verfügung, nur das eigene App-Fenster fehlt. '
            + 'Adresse: http://127.0.0.1:' + (port || 8765) + '/';
  if (el) el.textContent = msg;
  toast('Läuft im Standardbrowser (kein Edge/Chrome gefunden)', 'ok');
}

function compRow(c) {
  const detail = c.installed ? (c.version || c.path) : ('fehlt — ca. ' + c.download_mb + ' MB Download');
  // The go2rtc entry carries its running state with it, instead of in a
  // getrennten Block weiter unten.
  const sub = c.key === 'go2rtc' ? go2rtcSubBlock() : '';
  return '<div class="comp-item">'
    + '<span class="comp-dot ' + (c.installed ? 'ok' : 'missing') + '"></span>'
    + '<div class="comp-body">'
    + '<div class="comp-title">' + esc(c.label) + ' <span class="comp-purpose">— ' + esc(c.purpose) + '</span></div>'
    + '<div class="comp-detail">' + esc(detail) + '</div>'
    + sub
    + '</div></div>';
}

function go2rtcSubBlock() {
  return '<div class="comp-sub">'
    + '<div class="status-row">'
    + '<div class="sind checking" id="sind"></div>'
    + '<div class="sinfo"><h3 id="stitle"></h3><p id="sdesc"></p></div>'
    + '<div class="btn-inline">'
    + '<button class="hbtn" onclick="chkG2()" id="btn-check"></button>'
    + '<button class="hbtn green" onclick="restartG2()" id="btn-restart"></button>'
    + '<button class="hbtn red" onclick="killG2()" id="btn-killg2" title="Beendet alle go2rtc-Prozesse — auch verwaiste — und startet einen frischen">🧯 Alle go2rtc beenden</button>'
    + '</div></div>'
    + '<div class="comp-sub-row">'
    + '<span class="comp-sub-label">Prozesse:</span>'
    + '<span id="g2-proc" class="mono">–</span>'
    + '</div>'
    + '<div class="comp-sub-row">'
    + '<span class="comp-sub-label">Konfiguration:</span>'
    + '<a href="#" class="yaml-link" onclick="openDataFolder();return false;" id="yaml-path">go2rtc.yaml</a>'
    + '</div></div>';
}

async function loadComponents(showToast) {
  try {
    const r = await fetch('/api/components', {cache: 'no-store'});
    const d = await r.json();
    // Without Edge/Chrome the UI runs in the default browser — then the
    // app window is missing, but everything works.
    if (d.window_mode === 'browser') showBrowserHint(d.app_port);
    COMPONENTS = d.components || [];
    COMP_MISSING = d.missing || [];
    COMP_DOWNLOADABLE = !!d.downloadable;
    const list = document.getElementById('comp-list');
    if (list) {
      list.innerHTML = COMPONENTS.map(compRow).join('');
      applyLang();  // die neu erzeugten Knoepfe brauchen ihre Beschriftung
      chkG2();      // und ihren Laufzustand
      loadUpdateConfig();
    }
    if (showToast) toast(COMP_MISSING.length ? (COMP_MISSING.length + ' Komponente(n) fehlen') : 'Alle Komponenten vorhanden',
                         COMP_MISSING.length ? 'er' : 'ok');
    return d;
  } catch(e) { return null; }
}

// On start: only ask if something is really missing and the dialog is not
// dauerhaft weggeklickt wurde.
async function checkComponentsOnStart() {
  const d = await loadComponents(false);
  if (!d || d.dismissed) return;
  if ((d.missing || []).length) openSetupDialog();
}

function openSetupDialog() {
  const list = document.getElementById('setup-list');
  const missing = COMPONENTS.filter(c => !c.installed);
  if (!missing.length) { toast('Alle Komponenten sind vorhanden', 'ok'); return; }
  if (list) list.innerHTML = missing.map(compRow).join('');
  setShown(document.getElementById('setup-error'), false);
  setShown(document.getElementById('setup-progress'), false);
  const go = document.getElementById('setup-go');
  go.disabled = false;
  go.textContent = 'Jetzt herunterladen';
  if (!COMP_DOWNLOADABLE) {
    const err = document.getElementById('setup-error');
    setShown(err, true);
    err.textContent = 'Automatischer Download ist nur unter Windows x64 möglich. Lege go2rtc.exe in den Programmordner und ffmpeg.exe in den Unterordner ffmpeg\\.';
    go.disabled = true;
  }
  setShown(document.getElementById('setup-overlay'), true);
}

function closeSetupDialog() {
  setShown(document.getElementById('setup-overlay'), false);
  if (COMP_POLL) { clearInterval(COMP_POLL); COMP_POLL = null; }
}

async function dismissSetup(forever) {
  if (forever) {
    try { await fetch('/api/components/dismiss', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({forever:true})}); } catch(e) {}
    toast('Wird nicht mehr gefragt — nachholen im Reiter Setup', 'ok');
  }
  closeSetupDialog();
}

function fmtMB(bytes) { return (bytes / 1048576).toFixed(1) + ' MB'; }

async function startSetupDownload() {
  const go = document.getElementById('setup-go');
  go.disabled = true;
  go.textContent = 'Läuft…';
  setShown(document.getElementById('setup-error'), false);
  setShown(document.getElementById('setup-progress'), true);
  setText('setup-progress-text', 'Starte…');

  try {
    const r = await fetch('/api/components/install', {
      method: 'POST', headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({components: COMP_MISSING})
    });
    if (!r.ok) throw new Error((await r.text()).trim() || ('HTTP ' + r.status));
  } catch(e) {
    setupFailed(e.message || String(e));
    return;
  }

  if (COMP_POLL) clearInterval(COMP_POLL);
  COMP_POLL = setInterval(pollSetupProgress, 700);
}

function setupFailed(msg) {
  const err = document.getElementById('setup-error');
  setShown(err, true);
  err.textContent = msg;
  const go = document.getElementById('setup-go');
  go.disabled = false;
  go.textContent = 'Erneut versuchen';
  if (COMP_POLL) { clearInterval(COMP_POLL); COMP_POLL = null; }
}

const SETUP_STEPS = {download: 'Lade', extract: 'Entpacke', verify: 'Prüfe', done: 'Fertig', error: 'Fehler'};

async function pollSetupProgress() {
  let p;
  try { p = await (await fetch('/api/components/progress', {cache:'no-store'})).json(); }
  catch(e) { return; }

  const bar = document.getElementById('setup-bar');
  const txt = document.getElementById('setup-progress-text');
  const label = SETUP_STEPS[p.step] || p.step;

  if (p.total > 0) {
    const pct = Math.min(100, Math.round(p.received / p.total * 100));
    bar.style.setProperty('--w', pct + '%');
    txt.textContent = label + ' ' + (p.component || '') + ' — ' + fmtMB(p.received) + ' / ' + fmtMB(p.total) + ' (' + pct + '%)';
  } else {
    bar.style.setProperty('--w', p.step === 'download' ? '50%' : '85%');
    txt.textContent = label + ' ' + (p.component || '') + (p.received ? ' — ' + fmtMB(p.received) : '') + '…';
  }

  if (!p.done) return;
  clearInterval(COMP_POLL); COMP_POLL = null;

  if (p.error) { setupFailed(p.error); return; }
  bar.style.setProperty('--w', '100%');
  txt.textContent = 'Fertig — go2rtc wird gestartet…';
  await loadComponents(false);
  setTimeout(() => { closeSetupDialog(); chkG2(); toast('Komponenten installiert', 'ok'); }, 1200);
}

// ─── FS CSS ───────────────────────────────────────────────────────────────────

</script>
</body>
</html>`
