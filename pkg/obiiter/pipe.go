package obiiter

type Pipeable func(input IBioSequence) IBioSequence

func Pipeline(start Pipeable, parts ...Pipeable) Pipeable {
	p := func(input IBioSequence) IBioSequence {
		data := start(input)
		for _, part := range parts {
			data = part(data)
		}
		return data
	}

	return p
}

func (input IBioSequence) Pipe(start Pipeable, parts ...Pipeable) IBioSequence {
	p := Pipeline(start, parts...)
	return p(input)
}

type Teeable func(input IBioSequence) (IBioSequence, IBioSequence)

func (input IBioSequence) CopyTee() (IBioSequence, IBioSequence) {
	first := MakeIBioSequence()
	second := MakeIBioSequence()

	first.Add(1)

	go func() {
		first.WaitAndClose()
		second.Close()
	}()

	go func() {
		for input.Next() {
			b := input.Get()
			first.Push(b)
			second.Push(b)
		}
	}()

	return first, second
}
