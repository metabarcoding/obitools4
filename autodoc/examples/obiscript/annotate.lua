-- Adds a 'sample' attribute by extracting the prefix before the first underscore
function worker(sequence)
    local id = sequence:id()
    local sample = string.match(id, "^(.-)_")
    if sample then
        sequence:attribute("sample", sample)
    end
    return sequence
end
