# Gnutella 

## Données stockées par les noeuds
Chaque noeud disposera d'un entrepot sous forme de fichiers `.yaml`.

Ce fichier contiendra les metadonnées du contenu qui peut être partagé avec les autres noeuds.

Le fichier `.yaml` aura la structure suivante 

```yaml
- Title: ""
  Id: ""
  genre:
    - genre1
    - genre2
  stroage: "file_path"
```

## Algorithme distribué de decouverte 

On suppose que chaque noeud dispose de la liste de ces voisins.

### Structure de données au sein des noeuds:
Chaque noeud dispose d'un dictionaire clé, valeur. 

La clé correspond à l'identifiant unique d'une requete et la valeur l'adresse du noeud émeteur/initiateur de la requete. 

Une requete R1 initié par un noaud pour demander un fichier encapsule:
 - un Identifiant unique
 - le TTL (Time to live)
 - identifiant du document recherché

Une reponse P1 est générée par un noeud quand celui-ci possède le document.

P1 encapsule:
 - Une pile contenant le chemin parcouru. A chaque fois qu'un noeud recoit P1 il empile son adresse
 - l'identifiant unique provenant de la requete

### Algorithme
```
Un noeud initie une requete R1 de document et la transmet a ses voisins.
```

```
Si un noeud recoit R1 
    ajoute au dictionaire dict_emeteur["id_r1"] := adresse de l'emeteur
    Si le noeud dispose du document :
      Repond au noeud emeteur avec une reponse p1

    Decremente le TTL 
    Si TTL > 0
      Envoie la requete R1 a ses voisins
    Sinon 
      Arrete la propagation de R1

Sinon recoit P1
  empile l'adress du noeud actuel dans P1
  Pour chaque adresse dict_emeteur["id_r1]
    emet une reponse p

```