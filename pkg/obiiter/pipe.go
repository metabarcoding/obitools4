package obiiter


type Pipeable func(input IBioSequenceBatch) IBioSequenceBatch

func Pipeline(start Pipeable,parts ...Pipeable) Pipeable {
	p := func (input IBioSequenceBatch) IBioSequenceBatch {
		data := start(input)
		for _,part := range parts {
			data = part(data)
		}
		return data
	}

	return p
}

func (input IBioSequenceBatch) Pipe(start Pipeable, parts ...Pipeable) IBioSequenceBatch {
	p := Pipeline(start,parts...)
	return p(input)
}


type Teeable func(input IBioSequenceBatch) (IBioSequenceBatch,IBioSequenceBatch) 

func (input IBioSequenceBatch) CopyTee() (IBioSequenceBatch,IBioSequenceBatch) {
	first := MakeIBioSequenceBatch()
	second:= MakeIBioSequenceBatch()

	first.Add(1)

	go func() {
		first.WaitAndClose()
		second.Close()
	}()

	go func() {
		for input.Next() {
			b:=input.Get()
			first.Push(b)
			second.Push(b)
		}
	}()

	return first,second 
}
