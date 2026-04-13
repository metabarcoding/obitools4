-- Marks each sequence as processed by adding a 'processed' attribute
function worker(sequence)
    sequence:attribute("processed", "true")
    return sequence
end
