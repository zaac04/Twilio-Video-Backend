package mediaconvert

import (
	"context"
	"fmt"
	"log"
	AppConfig "stargazer/video-recording/config"
	"stargazer/video-recording/internal/enums"
	"stargazer/video-recording/internal/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert"
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert/types"
)

type MediaConvert struct {
	client         *mediaconvert.Client
	jobInput       []types.Input
	jobOutputGroup types.OutputGroup
	jobOutput      []types.Output
	RoleArn        string
	JobSettings    types.JobSettings
	outputLocation string
}

func GetMediaConvertEndpoint() string {
	return fmt.Sprintf("mediaconvert.%s.amazonaws.com", AppConfig.App.AWS_REGION)
}

func CreateClient() *MediaConvert {
	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	return &MediaConvert{
		client: mediaconvert.NewFromConfig(cfg),
	}
}

func (mc *MediaConvert) CreateJob(File []File, def []Definition, output_location string) (JobId string, err error) {

	mc.init(output_location)
	mc.addJobInput(File)
	mc.addOutputs(def)

	createJob := mediaconvert.CreateJobInput{
		Role:     utils.StringPtr(AppConfig.App.MEDIA_CONVERT_ROLE),
		Settings: &mc.JobSettings,
		UserMetadata: map[string]string{
			enums.ENV: AppConfig.App.ENV,
		},
	}

	out, err := mc.client.CreateJob(context.TODO(), &createJob)
	if err != nil {
		return "", fmt.Errorf("error occured while creating media convert job %v", err)
	}

	return *out.Job.Id, nil
}

func (mc *MediaConvert) init(output_location string) {
	mc.JobSettings = types.JobSettings{
		TimecodeConfig: &types.TimecodeConfig{
			Source: types.TimecodeSourceZerobased,
		},
	}
	mc.outputLocation = output_location
}

func (mc *MediaConvert) addJobInput(File []File) {
	for _, f := range File {
		mc.jobInput = append(mc.jobInput, types.Input{
			AudioSelectors: map[string]types.AudioSelector{
				"Audio Selector 1": {
					DefaultSelection:       types.AudioDefaultSelectionDefault,
					ExternalAudioFileInput: aws.String(f.AudioUrl),
				},
			},
			VideoSelector:  &types.VideoSelector{},
			TimecodeSource: types.InputTimecodeSourceZerobased,
			FileInput:      aws.String(f.VideoUrl),
		})
	}

	mc.JobSettings.Inputs = mc.jobInput
}

func (mc *MediaConvert) addOutputs(defs []Definition) {

	if len(defs) != 0 {

		mc.jobOutput = append(mc.jobOutput, types.Output{
			ContainerSettings: &types.ContainerSettings{
				Container: types.ContainerTypeMpd,
			},
			AudioDescriptions: []types.AudioDescription{
				{
					AudioSourceName: aws.String("Audio Selector 1"),
					CodecSettings: &types.AudioCodecSettings{
						Codec: types.AudioCodecAac,
						AacSettings: &types.AacSettings{
							Bitrate:    aws.Int32(128000),
							CodingMode: types.AacCodingModeCodingMode20,
							SampleRate: aws.Int32(48000),
						},
					},
				},
			},
			NameModifier: aws.String("_audio"),
		})

		for _, def := range defs {
			mc.jobOutput = append(mc.jobOutput, types.Output{

				ContainerSettings: &types.ContainerSettings{
					Container: types.ContainerTypeMpd,
				},
				VideoDescription: &types.VideoDescription{
					Width:  aws.Int32(ResolutionMap[def].width),
					Height: aws.Int32(ResolutionMap[def].height),
					CodecSettings: &types.VideoCodecSettings{
						Codec: types.VideoCodecH264,
						H264Settings: &types.H264Settings{
							Bitrate:         aws.Int32(ResolutionMap[def].bitrate),
							RateControlMode: types.H264RateControlModeCbr,
						},
					},
				},
				NameModifier: aws.String("_" + def.ToString()),
			})
		}

		mc.addToOutputGroup()
	}

}

func (mc *MediaConvert) addToOutputGroup() {
	mc.jobOutputGroup = types.OutputGroup{
		Outputs: mc.jobOutput,
		OutputGroupSettings: &types.OutputGroupSettings{
			Type: types.OutputGroupTypeDashIsoGroupSettings,
			DashIsoGroupSettings: &types.DashIsoGroupSettings{
				Destination:           aws.String(mc.outputLocation),
				SegmentLength:         aws.Int32(10),
				FragmentLength:        aws.Int32(2),
				MinBufferTime:         aws.Int32(2),
				SegmentControl:        types.DashIsoSegmentControlSegmentedFiles,
				MinFinalSegmentLength: aws.Float64(2),
			},
		},
	}
	mc.JobSettings.OutputGroups = append(mc.JobSettings.OutputGroups, mc.jobOutputGroup)

	mc.JobSettings.OutputGroups = append(mc.JobSettings.OutputGroups, types.OutputGroup{
		Outputs: []types.Output{
			{
				ContainerSettings: &types.ContainerSettings{
					Container: types.ContainerTypeMp4,
				},
				VideoDescription: &types.VideoDescription{
					Width:  aws.Int32(640),
					Height: aws.Int32(360),
					CodecSettings: &types.VideoCodecSettings{
						Codec: types.VideoCodecH264,
						H264Settings: &types.H264Settings{
							Bitrate:         aws.Int32(2500000),
							RateControlMode: types.H264RateControlModeCbr,
						},
					},
				},
				NameModifier: aws.String("_ml_output"),
			},
		},
		OutputGroupSettings: &types.OutputGroupSettings{
			Type: types.OutputGroupTypeFileGroupSettings,
			FileGroupSettings: &types.FileGroupSettings{
				Destination: aws.String(mc.outputLocation),
			},
		},
	})

}
