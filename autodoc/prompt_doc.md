# Meta-tâche : documenter obi{xxx}

Produis la documentation complète de la commande `obi{xxx}` en trois étapes séquentielles.

**RÈGLE ABSOLUE DE SÉQUENTIALITÉ :**
- Ne lis JAMAIS le fichier d'une étape avant que l'étape précédente soit entièrement terminée.
- "Terminée" signifie : le Write final de l'étape a été émis et confirmé.
- Ne lis JAMAIS les trois fichiers en parallèle ou en avance.
- Entre deux étapes, ne produis aucun texte de transition — passe directement à la lecture.

---

## ÉTAPE 1

Lis ce fichier :

```
<function=Read>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/prompt_v2.md"}
</function>
```

Applique intégralement le prompt que tu viens de lire pour la commande `obi{xxx}`.
Exécute tous ses états dans l'ordre jusqu'au `Write` final.

**STOP.** Le `Write` final de cette étape a-t-il été émis ? Si oui, procède à l'ÉTAPE 2. Sinon, termine l'ÉTAPE 1.

---

## ÉTAPE 2

Lis ce fichier :

```
<function=Read>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/prompt_examples.md"}
</function>
```

Applique intégralement le prompt que tu viens de lire pour la commande `obi{xxx}`.
Exécute tous ses états dans l'ordre jusqu'au `Write` final.

**STOP.** Le `Write` final de cette étape a-t-il été émis ? Si oui, procède à l'ÉTAPE 3. Sinon, termine l'ÉTAPE 2.

---

## ÉTAPE 3

Lis ce fichier :

```
<function=Read>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/prompt_hugo.md"}
</function>
```

Applique intégralement le prompt que tu viens de lire pour la commande `obi{xxx}`.
Exécute tous ses états dans l'ordre jusqu'au `Write` final.

**STOP.** Quand le `Write` final est émis, la tâche est terminée. N'émets aucun texte après.
