# Erzeugt ein COFF-Objekt (.syso) mit Icon-Ressourcen. Der Go-Linker bindet
# eine Datei mit dieser Endung automatisch ein — damit traegt die Programmdatei
# selbst ein Symbol. Das ist die Ebene, die Windows beim Start zeigt, noch bevor
# ein Browser oder ein Favicon ueberhaupt eine Rolle spielt.
import struct, sys

RT_ICON, RT_GROUP_ICON = 3, 14

def dir_hdr(named, ids):
    return struct.pack('<IIHHHH', 0, 0, 0, 0, named, ids)

def dir_entry(ident, offset, ist_verzeichnis):
    if ist_verzeichnis: offset |= 0x80000000
    return struct.pack('<II', ident, offset)

def baue(ico_pfad, syso_pfad):
    roh = open(ico_pfad, 'rb').read()
    _, typ, anzahl = struct.unpack('<HHH', roh[:6])
    assert typ == 1, 'keine ICO-Datei'
    bilder = []
    for i in range(anzahl):
        e = roh[6 + 16*i : 6 + 16*(i+1)]
        b, h, farben, _, ebenen, bits, groesse, offset = struct.unpack('<BBBBHHII', e)
        bilder.append(dict(b=b, h=h, farben=farben, ebenen=ebenen, bits=bits,
                           daten=roh[offset:offset+groesse]))

    # GROUP_ICON-Verzeichnis: verweist per Kennung auf die einzelnen Symbole
    grp = struct.pack('<HHH', 0, 1, len(bilder))
    for i, b in enumerate(bilder, start=1):
        grp += struct.pack('<BBBBHHIH', b['b'], b['h'], b['farben'], 0,
                           b['ebenen'], b['bits'], len(b['daten']), i)

    # Blaetter: (Typ, Kennung, Daten)
    blaetter = [(RT_ICON, i, b['daten']) for i, b in enumerate(bilder, start=1)]
    blaetter.append((RT_GROUP_ICON, 1, grp))

    typen = {}
    for t, kennung, daten in blaetter:
        typen.setdefault(t, []).append((kennung, daten))

    # Groessen vorausberechnen, damit die Verweise stimmen
    n_typen = len(typen)
    groesse_wurzel = 16 + 8*n_typen
    groesse_namen = sum(16 + 8*len(v) for v in typen.values())
    groesse_sprachen = sum(16 + 8 for v in typen.values() for _ in v)
    anzahl_blaetter = sum(len(v) for v in typen.values())
    off_datenverzeichnis = groesse_wurzel + groesse_namen + groesse_sprachen
    off_rohdaten = off_datenverzeichnis + 16*anzahl_blaetter

    wurzel = dir_hdr(0, n_typen)
    namen_bloecke, sprach_bloecke, daten_eintraege, rohdaten = b'', b'', b'', b''
    off_namen = groesse_wurzel
    off_sprachen = groesse_wurzel + groesse_namen
    off_daten = off_datenverzeichnis
    off_roh = off_rohdaten
    relokationen = []

    for t in sorted(typen):
        wurzel += dir_entry(t, off_namen + len(namen_bloecke), True)

    for t in sorted(typen):
        eintraege = typen[t]
        namen_bloecke += dir_hdr(0, len(eintraege))
        for kennung, daten in eintraege:
            namen_bloecke += dir_entry(kennung, off_sprachen + len(sprach_bloecke), True)
            sprach_bloecke += dir_hdr(0, 1)
            sprach_bloecke += dir_entry(0x0409, off_daten + len(daten_eintraege), False)
            # Das Feld OffsetToData wird vom Linker verschoben — dafuer die Relokation
            relokationen.append(off_daten + len(daten_eintraege))
            daten_eintraege += struct.pack('<IIII', off_roh + len(rohdaten), len(daten), 0, 0)
            rohdaten += daten
            if len(rohdaten) % 8:
                rohdaten += b'\x00' * (8 - len(rohdaten) % 8)

    abschnitt = wurzel + namen_bloecke + sprach_bloecke + daten_eintraege + rohdaten

    # COFF zusammensetzen
    anzahl_reloc = len(relokationen)
    off_abschnittsdaten = 20 + 40
    off_reloc = off_abschnittsdaten + len(abschnitt)
    off_symbole = off_reloc + 10*anzahl_reloc

    kopf = struct.pack('<HHIIIHH', 0x8664, 1, 0, off_symbole, 2, 0, 0)
    abschnittskopf = (b'.rsrc\x00\x00\x00' +
        struct.pack('<IIIIIIHHI', 0, 0, len(abschnitt), off_abschnittsdaten,
                    off_reloc if anzahl_reloc else 0, 0, anzahl_reloc, 0, 0x40000040))

    reloc = b''
    for adresse in relokationen:
        # IMAGE_REL_AMD64_ADDR32NB = 3, Symbol 0 = der Abschnitt selbst
        reloc += struct.pack('<IIH', adresse, 0, 3)

    # Ein COFF-Symbol ist GENAU 18 Byte, der Zusatzeintrag ebenso. Bei 20 Byte
    # verliert der Linker den Faden und verwirft die Ressource stillschweigend —
    # der Bau laeuft durch, das Symbol fehlt trotzdem.
    sym = struct.pack('<8sIhHBB', b'.rsrc\x00\x00\x00', 0, 1, 0, 3, 1)
    assert len(sym) == 18, len(sym)
    aux = struct.pack('<IHHIHB', len(abschnitt), anzahl_reloc, 0, 0, 0, 0) + b'\x00'*3
    assert len(aux) == 18, len(aux)
    symbole = sym + aux
    strings = struct.pack('<I', 4)

    open(syso_pfad, 'wb').write(kopf + abschnittskopf + abschnitt + reloc + symbole + strings)
    print('%s: %d Bytes, %d Symbole, %d Relokationen' %
          (syso_pfad, len(abschnitt), len(bilder), anzahl_reloc))

baue(sys.argv[1], sys.argv[2])
