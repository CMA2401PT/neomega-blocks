package ir_go

// import nbt "neomega-nbt"

// func (b IRGoBlock) DecodeToHighLevel(encoder nbt.EncoderHighLevel) (err error) {
// 	err = encoder.WriteRootTag("", nbt.ComplexCompound)
// 	if err != nil {
// 		return err
// 	}
// 	err = encoder.BeginComplexCompound()
// 	if err != nil {
// 		return err
// 	}
// 	err = encoder.WriteCompoundTag("name", nbt.TrivialString)
// 	if err != nil {
// 		return err
// 	}
// 	err = encoder.WriteString(b.nameWithNameSpace, nbt.ValuePosCompoundOrRoot)
// 	if err != nil {
// 		return err
// 	}
// 	err = encoder.NextCompoundElement()
// 	if err != nil {
// 		return err
// 	}
// 	err = encoder.WriteCompoundTag("states", nbt.ComplexCompound)
// 	if err != nil {
// 		return err
// 	}
// 	err = encoder.BeginComplexCompound()
// 	if err != nil {
// 		return err
// 	}

// 	for i, prop := range b.props {
// 		if i != 0 {
// 			err = encoder.NextCompoundElement()
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		if prop.Value.IsInt8() {
// 			err = encoder.WriteCompoundTag(prop.Key, nbt.TrivialByte)
// 			if err != nil {
// 				return err
// 			}
// 			err = encoder.WriteInt8(prop.Value.AsInt8(), nbt.ValuePosCompoundOrRoot)
// 			if err != nil {
// 				return err
// 			}
// 		} else if prop.Value.IsInt32() {
// 			err = encoder.WriteCompoundTag(prop.Key, nbt.TrivialInt)
// 			if err != nil {
// 				return err
// 			}
// 			err = encoder.WriteInt32(prop.Value.AsInt32(), nbt.ValuePosCompoundOrRoot)
// 			if err != nil {
// 				return err
// 			}
// 		} else {
// 			err = encoder.WriteCompoundTag(prop.Key, nbt.TrivialString)
// 			if err != nil {
// 				return err
// 			}
// 			err = encoder.WriteString(prop.Value.AsString(), nbt.ValuePosCompoundOrRoot)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	err = encoder.EndComplexCompound()
// 	if err != nil {
// 		return err
// 	}
// 	return encoder.EndComplexCompound()
// }
