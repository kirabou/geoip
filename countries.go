package geoip



import (
	"fmt"
	"encoding/csv"
	"strings"
	"github.com/google/btree"
	"io"
)



// Country structure holds information for a given country :
// its 2 characters code (ISO 3166-1 alpha 2), and its name.
// Example : 
// 	{ "FR", "France" }
type Country struct {
	Code string	
	Name string
}


// Countries type is a BTree of Country structures
type Countries btree.BTree


// Implements String() function to *Country type, so it
// implements the Stringer interface an can be Println()
func (country *Country) String() string {
	return fmt.Sprintf("Counttry.Code=%q, Country.Name=%q",
		country.Code, country.Name)
}


// Implements the Item interface from btree package for
// the Country type, so we can use them in a btree.
// Less tests whether the current item is less than the given argument.
func (country Country)Less(than btree.Item) bool {
	return country.Code < than.(Country).Code
}


// LoadCountries() loads the countries (as defined in the local 
// countries constant) in a memory BTree
func LoadCountries() (*Countries, error) {

	r := csv.NewReader(strings.NewReader(countries_list))
	r.FieldsPerRecord = -1
    r.Comma = ';'
    
    t := btree.New(4)
    
    for {

    	values, err := r.Read()
    	if err == io.EOF {
    		break
    	}
    	if err != nil {
    		log_geolocip.Err(fmt.Sprintf("Countries error %v", err))
    		break
    	}

    	// fmt.Println(len(values), values)
	
		// Use only lines with 2 values
	   	if len(values) == 2 {
	   		t.ReplaceOrInsert(Country{ values[1], values[0] })
	   	}
    }

    return (*Countries)(t), nil
}


// Get() returns the Country structure matching a given country code
func (countries *Countries)Get(country_code string) *Country {
	tree := (*btree.BTree)(countries)
	item := tree.Get(Country{country_code, ""})
	if item != nil {
		country := item.(Country)
		return(&country)
	} else {
		return(nil)
	}
}


// CSV list of country names and ISO3661 codes
const (
	countries_list = `Afghanistan;AF
Albanie;AL
Algérie;DZ
Samoa Américaines;AS
Andorre;AD
Angola;AO
Anguilla;AI
Antarctique;AQ
Antigua-Et-Barbuda;AG
Argentine;AR
Arménie;AM
Aruba;AW
Australie;AU
Autriche;AT
Azerbaïdjan;AZ
Bahamas;BS
Bahreïn;BH
Bangladesh;BD
Barbade;BB
Bélarus;BY
Belgique;BE
Belize;BZ
Bénin;BJ
Bermudes;BM
Bhoutan;BT
Bolivie, l'État Plurinational de;BO
Bonaire, Saint-Eustache et Saba;BQ
Bosnie-Herzégovine;BA
Botswana;BW
Bouvet, Île;BV
Brésil;BR
Océan Indien, Territoire Britannique de l';IO
Brunei Darussalam;BN
Bulgarie;BG
Burkina Faso;BF
Burundi;BI
Cambodge;KH
Cameroun;CM
Canada;CA
Cap-Vert;CV
Caïmans, Îles;KY
Centrafricaine, République;CF
Tchad;TD
Chili;CL
Chine;CN
Christmas, Île;CX
Cocos (Keeling), Îles;CC
Colombie;CO
Comores;KM
Congo;CG
Congo, la République Démocratique du;CD
Cook, Îles;CK
Costa Rica;CR
Croatie;HR
Cuba;CU
Curaçao;CW
Chypre;CY
Tchèque, République;CZ
Côte d'Ivoire;CI
Danemark;DK
Djibouti;DJ
Dominique;DM
Dominicaine, République;DO
Équateur;EC
Égypte;EG
El Salvador;SV
Guinée Équatoriale;GQ
Érythrée;ER
Estonie;EE
Éthiopie;ET
Falkland, Îles (Malvinas);FK
Féroé, Îles;FO
Fidji;FJ
Finlande;FI
France;FR
Guyane Française;GF
Polynésie Française;PF
Terres Australes Françaises;TF
Gabon;GA
Gambie;GM
Géorgie;GE
Allemagne;DE
Ghana;GH
Gibraltar;GI
Grèce;GR
Groenland;GL
Grenade;GD
Guadeloupe;GP
Guam;GU
Guatemala;GT
Guernesey;GG
Guinée;GN
Guinée-Bissau;GW
Guyana;GY
Haïti;HT
Heard-Et-Îles Macdonald, Île;HM
Saint-Siège (État de la Cité du Vatican);VA
Honduras;HN
Hong Kong;HK
Hongrie;HU
Islande;IS
Inde;IN
Indonésie;ID
Iran, République Islamique d';IR
Iraq;IQ
Irlande;IE
Île de Man;IM
Israël;IL
Italie;IT
Jamaïque;JM
Japon;JP
Jersey;JE
Jordanie;JO
Kazakhstan;KZ
Kenya;KE
Kiribati;KI
Corée, République Populaire Démocratique de;KP
Corée, République de;KR
Koweït;KW
Kirghizistan;KG
Lao, République Démocratique Populaire;LA
Lettonie;LV
Liban;LB
Lesotho;LS
Libéria;LR
Libye;LY
Liechtenstein;LI
Lituanie;LT
Luxembourg;LU
Macao;MO
Macédoine, l'Ex-république Yougoslave de;MK
Madagascar;MG
Malawi;MW
Malaisie;MY
Maldives;MV
Mali;ML
Malte;MT
Marshall, Îles;MH
Martinique;MQ
Mauritanie;MR
Maurice;MU
Mayotte;YT
Mexique;MX
Micronésie, États Fédérés de;FM
Moldova, République de;MD
Monaco;MC
Mongolie;MN
Monténégro;ME
Montserrat;MS
Maroc;MA
Mozambique;MZ
Myanmar;MM
Namibie;NA
Nauru;NR
Népal;NP
Pays-Bas;NL
Nouvelle-Calédonie;NC
Nouvelle-Zélande;NZ
Nicaragua;NI
Niger;NE
Nigéria;NG
Niué;NU
Norfolk, Île;NF
Mariannes du Nord, Îles;MP
Norvège;NO
Oman;OM
Pakistan;PK
Palaos;PW
Palestine, État de;PS
Panama;PA
Papouasie-Nouvelle-Guinée;PG
Paraguay;PY
Pérou;PE
Philippines;PH
Pitcairn;PN
Pologne;PL
Portugal;PT
Porto Rico;PR
Qatar;QA
Roumanie;RO
Russie, Fédération de;RU
Rwanda;RW
Réunion;RE
Saint-Barthélemy;BL
Sainte-Hélène, Ascension et Tristan da Cunha;SH
Saint-Kitts-Et-Nevis;KN
Sainte-Lucie;LC
Saint-Martin(partie Française);MF
Saint-Pierre-Et-Miquelon;PM
Saint-Vincent-Et-Les Grenadines;VC
Samoa;WS
Saint-Marin;SM
Sao Tomé-Et-Principe;ST
Arabie Saoudite;SA
Sénégal;SN
Serbie;RS
Seychelles;SC
Sierra Leone;SL
Singapour;SG
Saint-Martin (Partie Néerlandaise);SX
Slovaquie;SK
Slovénie;SI
Salomon, Îles;SB
Somalie;SO
Afrique du Sud;ZA
Géorgie du Sud-Et-Les Îles Sandwich du Sud;GS
Soudan du Sud;SS
Espagne;ES
Sri Lanka;LK
Soudan;SD
Suriname;SR
Svalbard et Île Jan Mayen;SJ
Swaziland;SZ
Suède;SE
Suisse;CH
Syrienne, République Arabe;SY
Taïwan, Province de Chine;TW
Tadjikistan;TJ
Tanzanie, République-Unie de;TZ
Thaïlande;TH
Timor-Leste;TL
Togo;TG
Tokelau;TK
Tonga;TO
Trinité-Et-Tobago;TT
Tunisie;TN
Turquie;TR
Turkménistan;TM
Turks-Et-Caïcos, Îles;TC
Tuvalu;TV
Ouganda;UG
Ukraine;UA
Émirats Arabes Unis;AE
Royaume-Uni;GB
États-Unis;US
Îles Mineures Éloignées des États-Unis;UM
Uruguay;UY
Ouzbékistan;UZ
Vanuatu;VU
Venezuela, République Bolivarienne du;VE
Viet Nam;VN
Îles Vierges Britanniques;VG
Îles Vierges des États-Unis;VI
Wallis et Futuna;WF
Sahara Occidental;EH
Yémen;YE
Zambie;ZM
Zimbabwe;ZW`
)

