# Guide de rédaction d'un obitest

## Règles essentielles

1. **Données < 1 KB** - Fichiers de test très petits
2. **Exécution < 10 sec** - Tests rapides pour CI/CD
3. **Auto-contenu** - Pas de dépendances externes
4. **Auto-nettoyage** - Pas de fichiers résiduels

## Structure minimale

```
obitests/obitools/<commande>/
├── test.sh          # Script exécutable
└── data.fasta       # Données minimales (optionnel)
```

## Template de test.sh

```bash
#!/bin/bash

TEST_NAME=<commande>
CMD=<commande>

TEST_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
OBITOOLS_DIR="${TEST_DIR/obitest*/}build"
export PATH="${OBITOOLS_DIR}:${PATH}"

MCMD="$(echo "${CMD:0:4}" | tr '[:lower:]' '[:upper:]')$(echo "${CMD:4}" | tr '[:upper:]' '[:lower:]')"

TMPDIR="$(mktemp -d)"
ntest=0
success=0
failed=0

cleanup() {
    echo "========================================" 1>&2
    echo "## Results of the $TEST_NAME tests:" 1>&2
    echo 1>&2
    echo "- $ntest tests run" 1>&2
    echo "- $success successfully completed" 1>&2
    echo "- $failed failed tests" 1>&2
    echo 1>&2
    echo "Cleaning up the temporary directory..." 1>&2
    echo 1>&2
    echo "========================================" 1>&2

    rm -rf "$TMPDIR"

    if [ $failed -gt 0 ]; then
       log "$TEST_NAME tests failed" 
        log
        log
       exit 1
    fi

    log
    log
    exit 0
}

log() {
    echo -e "[$TEST_NAME @ $(date)] $*" 1>&2
}

log "Testing $TEST_NAME..." 
log "Test directory is $TEST_DIR" 
log "obitools directory is $OBITOOLS_DIR" 
log "Temporary directory is $TMPDIR" 
log "files: $(find $TEST_DIR | awk -F'/' '{print $NF}' | tail -n +2)"

########## TESTS ##########

# Test 1: Help (OBLIGATOIRE)
((ntest++))
if $CMD -h > "${TMPDIR}/help.txt" 2>&1 
then
    log "$MCMD: printing help OK" 
    ((success++))
else
    log "$MCMD: printing help failed" 
    ((failed++))
fi

# Ajoutez vos tests ici...

###########################

cleanup
```

## Pattern de test

```bash
((ntest++))
if commande args > "${TMPDIR}/output.txt" 2>&1
then
    log "$MCMD: description OK" 
    ((success++))
else
    log "$MCMD: description failed"
    ((failed++))
fi
```

## Tests courants

### Exécution basique
```bash
((ntest++))
if $CMD "${TEST_DIR}/input.fasta" > "${TMPDIR}/output.fasta" 2>&1
then
    log "$MCMD: basic execution OK" 
    ((success++))
else
    log "$MCMD: basic execution failed"
    ((failed++))
fi
```

### Sortie non vide
```bash
((ntest++))
if [ -s "${TMPDIR}/output.fasta" ]
then
    log "$MCMD: output not empty OK"
    ((success++))
else
    log "$MCMD: output empty - failed"
    ((failed++))
fi
```

### Comptage
```bash
((ntest++))
count=$(grep -c "^>" "${TMPDIR}/output.fasta")
if [ "$count" -gt 0 ]
then
    log "$MCMD: extracted $count sequences OK"
    ((success++))
else
    log "$MCMD: no sequences - failed"
    ((failed++))
fi
```

### Présence de contenu
```bash
((ntest++))
if grep -q "expected_string" "${TMPDIR}/output.fasta"
then
    log "$MCMD: expected content found OK"
    ((success++))
else
    log "$MCMD: content not found - failed"
    ((failed++))
fi
```

### Comparaison avec référence
```bash
((ntest++))
if diff "${TEST_DIR}/expected.fasta" "${TMPDIR}/output.fasta" > /dev/null
then
    log "$MCMD: matches reference OK"
    ((success++))
else
    log "$MCMD: differs from reference - failed"
    ((failed++))
fi
```

### Test avec options
```bash
((ntest++))
if $CMD --opt value "${TEST_DIR}/input.fasta" > "${TMPDIR}/out.fasta" 2>&1
then
    log "$MCMD: with option OK" 
    ((success++))
else
    log "$MCMD: with option failed"
    ((failed++))
fi
```

## Variables importantes

- **TEST_DIR** - Répertoire du test (données d'entrée)
- **TMPDIR** - Répertoire temporaire (sorties)
- **CMD** - Nom de la commande
- **MCMD** - Nom formaté pour les logs

## Règles d'or

✅ **Entrées** → `${TEST_DIR}/`
✅ **Sorties** → `${TMPDIR}/`
✅ **Toujours rediriger** → `> file 2>&1`
✅ **Incrémenter ntest** → Avant chaque test
✅ **Messages clairs** → Descriptions explicites

❌ **Pas de chemins en dur**
❌ **Pas de /tmp direct**
❌ **Pas de sortie vers TEST_DIR**
❌ **Pas de commandes sans redirection**

## Données de test

Créer un fichier minimal (< 500 bytes) :

```fasta
>seq1
ACGTACGTACGTACGT
>seq2
AAAACCCCGGGGTTTT
>seq3
ATCGATCGATCGATCG
```

## Création rapide

```bash
# 1. Créer le répertoire
mkdir -p obitests/obitools/<commande>
cd obitests/obitools/<commande>

# 2. Créer les données de test
cat > test_data.fasta << 'EOF'
>seq1
ACGTACGTACGTACGT
>seq2
AAAACCCCGGGGTTTT
EOF

# 3. Copier le template dans test.sh
# 4. Adapter le TEST_NAME et CMD
# 5. Ajouter les tests
# 6. Rendre exécutable
chmod +x test.sh

# 7. Tester
./test.sh
```

## Checklist

- [ ] `test.sh` exécutable (`chmod +x`)
- [ ] Test d'aide inclus
- [ ] Données < 1 KB
- [ ] Sorties vers `${TMPDIR}/`
- [ ] Entrées depuis `${TEST_DIR}/`
- [ ] Redirections `2>&1`
- [ ] Messages clairs
- [ ] Testé localement
- [ ] Exit code 0 si succès

## Debug

Conserver TMPDIR pour inspection :
```bash
cleanup() {
    echo "Temporary directory: $TMPDIR" 1>&2
    # rm -rf "$TMPDIR"  # Commenté
    ...
}
```

Mode verbose :
```bash
set -x  # Au début du script
```

## Exemples

**Simple (1 test)** - obimicrosat
```bash
# Juste l'aide
```

**Moyen (4-5 tests)** - obisuperkmer
```bash
# Aide + exécution + validation sortie + contenu
```

**Complet (7+ tests)** - obiuniq
```bash
# Aide + exécution + comparaison CSV + options + multiples cas
```

## Commandes utiles

```bash
# Compter séquences
grep -c "^>" file.fasta

# Fichier non vide
[ -s file ]

# Comparer
diff file1 file2 > /dev/null

# Comparer compressés
zdiff file1.gz file2.gz

# Compter bases
grep -v "^>" file | tr -d '\n' | wc -c
```

## Ce qu'il faut retenir

Un bon test est **COURT**, **RAPIDE** et **SIMPLE** :
- 3-10 tests maximum
- Données < 1 KB
- Exécution < 10 secondes
- Pattern standard respecté
