package ir

import (
	decoder_bedrock_block_props "neomega-blocks/encode-decode/decoder-bedrock-block-props"
	nbt "neomega-nbt"
)

func decodeToHL[E nbt.EncoderHighLevel](b IRBlock, encoder E) (err error) {
	err = encoder.WriteRootTag("", nbt.ComplexCompound)
	if err != nil {
		return err
	}
	err = encoder.BeginComplexCompound()
	if err != nil {
		return err
	}
	err = encoder.WriteCompoundTag("name", nbt.TrivialString)
	if err != nil {
		return err
	}
	err = encoder.WriteString(b.NameWithNameSpace(), nbt.ValuePosCompoundOrRoot)
	if err != nil {
		return err
	}
	err = encoder.NextCompoundElement()
	if err != nil {
		return err
	}
	err = encoder.WriteCompoundTag("states", nbt.ComplexCompound)
	if err != nil {
		return err
	}
	err = decoder_bedrock_block_props.Decode(b.States(), encoder, false)
	if err != nil {
		return err
	}
	return encoder.EndComplexCompound()
}
