function begin()
    obicontext.item("compteur", 0)
end

function worker(sequence)
    samples = sequence:attribute("merged_sample")
    samples["tutu"] = 4
    sequence:attribute("merged_sample", samples)
    sequence:attribute("toto", 44444)
    nb = obicontext.inc("compteur")
    sequence:id("seq_" .. nb)
    return sequence
end

function finish()
    print("compteur = " .. obicontext.item("compteur"))
end
